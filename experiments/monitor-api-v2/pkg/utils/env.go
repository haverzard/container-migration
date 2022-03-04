package utils

import (
	"os"
	"strconv"
)

var (
	TIME_WEIGHT            float64 = 100
	METRIC_WEIGHT          float64 = 20
	PROGRESSING_THREESHOLD float64 = 10
	CONVERGED_THREESHOLD   float64 = -25
	OVERLOAD_THREESHOLD    float64 = 0.2
	SERVER_ENDPOINT        string  = ""
	NODE_NAME              string  = ""
)

func LoadEnv() {
	var t string // temporary variable
	t = os.Getenv("TIME_WEIGHT")
	if t != "" {
		if res, err := strconv.ParseFloat(t, 64); err == nil {
			TIME_WEIGHT = res
		}
	}
	t = os.Getenv("METRIC_WEIGHT")
	if t != "" {
		if res, err := strconv.ParseFloat(t, 64); err == nil {
			METRIC_WEIGHT = res
		}
	}
	t = os.Getenv("PROGRESSING_THREESHOLD")
	if t != "" {
		if res, err := strconv.ParseFloat(t, 64); err == nil {
			PROGRESSING_THREESHOLD = res
		}
	}
	t = os.Getenv("CONVERGED_THREESHOLD")
	if t != "" {
		if res, err := strconv.ParseFloat(t, 64); err == nil {
			CONVERGED_THREESHOLD = res
		}
	}
	t = os.Getenv("OVERLOAD_THREESHOLD")
	if t != "" {
		if res, err := strconv.ParseFloat(t, 64); err == nil {
			OVERLOAD_THREESHOLD = res
		}
	}
	SERVER_ENDPOINT = os.Getenv("SERVER_ENDPOINT")
	NODE_NAME = os.Getenv("NODE_NAME")
}
