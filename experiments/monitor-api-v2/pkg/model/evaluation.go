package model

import "time"

type EvaluationObject struct {
	Pod   string  `json:"pod"`
	Value float64 `json:"value"`
}

type PodEvaluation struct {
	Metric float64
	Time   time.Time
}
