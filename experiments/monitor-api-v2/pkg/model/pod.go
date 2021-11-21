package model

import (
	"log"
	"time"

	"github.com/haverzard/ta/pkg/utils"
	"github.com/oleiade/lane"
)

type Pod struct {
	Name        string
	Score       float64
	AccessedAt  time.Time
	Category    PodCategory
	Evaluations *lane.Queue
}

func NewPod(name string) *Pod {
	return &Pod{Name: name, Score: 0, AccessedAt: time.Now(), Category: Progressing, Evaluations: lane.NewQueue()}
}

func (pod *Pod) Speculate() {
	// Speculate if there are 2 evaluation results
	n := pod.Evaluations.Size()
	if n < 2 {
		return
	}

	// Only speculate using latest two
	for ; n > 2; n-- {
		pod.Evaluations.Pop()
	}
	oldest := pod.Evaluations.Pop().(*PodEvaluation)
	newest := pod.Evaluations.First().(*PodEvaluation)

	// Find differences
	dt := float64((newest.Time.UnixNano() - oldest.Time.UnixNano()) / 1000000)
	dmetric := newest.Metric - oldest.Metric

	// Get score
	score := (dmetric*utils.METRIC_WEIGHT - dt/utils.TIME_WEIGHT)

	oldCategory := pod.Category
	dscore := score - pod.Score

	log.Printf("Time: %v, Score: %v, dScore: %v", dt, score, dscore)
	log.Printf("Test %v", utils.PROGRESSING_THREESHOLD)
	// SpeCon + Custom categorization
	if dscore > utils.PROGRESSING_THREESHOLD || dscore > float64(5) {
		pod.Category = Progressing
	} else if dscore < float64(-5) {
		if pod.Category == Progressing && dscore >= utils.CONVERGED_THREESHOLD {
			pod.Category = Watching
		} else {
			pod.Category = Converged
		}
	}
	if oldCategory != pod.Category {
		pod.Score = score
	}
}

func (pod *Pod) AddEvaluation(eval *EvaluationObject) {
	pod.Evaluations.Enqueue(&PodEvaluation{Metric: eval.Value, Time: time.Now()})
}
