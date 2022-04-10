package utils

import (
	"os"
	"strconv"
)

var (
	OVERLOAD_THREESHOLD float64 = 1.0
	SERVER_ENDPOINT     string  = ""
	NODE_NAME           string  = ""
)

func LoadEnv() {
	// temporary variable
	t := os.Getenv("OVERLOAD_THREESHOLD")
	if t != "" {
		if res, err := strconv.ParseFloat(t, 64); err == nil {
			OVERLOAD_THREESHOLD = res
		}
	}
	SERVER_ENDPOINT = os.Getenv("SERVER_ENDPOINT")
	NODE_NAME = os.Getenv("NODE_NAME")
}
