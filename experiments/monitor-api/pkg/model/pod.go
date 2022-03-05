package model

import (
	"fmt"
	"log"
	"math"
	"sync"
	"time"

	"github.com/oleiade/lane"
	"github.com/sajari/regression"
)

type Pod struct {
	Name          string
	Score         float64
	CreatedAt     time.Time
	LastMigration time.Time
	AccessedAt    time.Time
	Category      PodCategory
	Evaluations   *lane.Queue
	Regressor     *regression.Regression
	mu            sync.Mutex
}

func NewPod(name string) *Pod {
	t := time.Now()
	r := new(regression.Regression)
	r.SetObserved("log y")
	r.SetVar(0, "t")
	return &Pod{
		Name:        name,
		Score:       0,
		CreatedAt:   t,
		AccessedAt:  t,
		Category:    Progressing,
		Evaluations: lane.NewQueue(),
		Regressor:   r,
	}
}

func (pod *Pod) Speculate() {
	// Speculate if there are 2 evaluation results
	n := pod.Evaluations.Size()
	if n < 2 {
		return
	}

	// Only speculate using latest two
	pod.mu.Lock()
	for ; n > 2; n-- {
		pod.Evaluations.Pop()
	}
	eval1 := pod.Evaluations.Pop().(*PodEvaluation)
	eval2 := pod.Evaluations.First().(*PodEvaluation)
	pod.mu.Unlock()

	// Find differences
	t1 := float64(eval1.Time.UnixMilli())
	t2 := float64(eval2.Time.UnixMilli())
	t0 := float64(pod.CreatedAt.UnixMilli())
	// dt := float64((newest.Time.UnixNano() + oldest.Time.UnixNano()) / 1000000)

	// Get score
	score := math.Abs((eval2.Metric - eval1.Metric) / (t2 - t1))
	alpha := float64(0.0)
	beta := float64(0.0)

	pod.Regressor.Train(regression.DataPoint(math.Log(score)/math.Log(math.E), []float64{t2 - t0}))
	if err := pod.Regressor.Run(); err == nil {
		alpha, err = pod.Regressor.Predict([]float64{t2 - t0})
		if err != nil {
			log.Fatalln(err)
			return
		}
		alpha = math.Exp(alpha)
		fmt.Println(alpha)
		beta = 1 - pod.Regressor.R2
	}

	// log.Printf("Time: %v, Score: %v, dScore: %v", dt, score, dscore)
	// log.Printf("Test %v", utils.PROGRESSING_THREESHOLD)
	// SpeCon + Custom categorization
	if score > alpha {
		pod.Category = Progressing
	} else {
		if pod.Category == Progressing {
			pod.Category = Watching
		} else if score+score*beta < pod.Score {
			pod.Category = Converged
		}
	}
	pod.Score = score
}

func (pod *Pod) AddEvaluation(eval *EvaluationObject) {
	pod.mu.Lock()
	pod.Evaluations.Enqueue(&PodEvaluation{Metric: eval.Value, Time: time.Now()})
	pod.mu.Unlock()
}
