package model

import (
	"time"

	"github.com/haverzard/ta/pkg/utils"
	"github.com/oleiade/lane"
)

type Pod struct {
	Name        string
	Score       float64
	NextRetry   int64
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
	t1 := float64(oldest.Time.UnixNano() / 1000000)
	t2 := float64(newest.Time.UnixNano() / 1000000)
	// dt := float64((newest.Time.UnixNano() + oldest.Time.UnixNano()) / 1000000)

	// How many percentage do the metric & time change?
	time_score := (t2 - t1) / t1
	metric_score := (newest.Metric - oldest.Metric) / oldest.Metric

	// Get score
	score := (metric_score - time_score)

	oldCategory := pod.Category
	dscore := score - pod.Score

	// log.Printf("Time: %v, Score: %v, dScore: %v", dt, score, dscore)
	// log.Printf("Test %v", utils.PROGRESSING_THREESHOLD)
	// SpeCon + Custom categorization
	if dscore > utils.PROGRESSING_THREESHOLD {
		pod.Category = Progressing
	} else if dscore < utils.CONVERGED_THREESHOLD {
		if pod.Category == Progressing {
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
