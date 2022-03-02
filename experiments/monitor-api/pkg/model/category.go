package model

type PodCategory int64

const (
	// since iota starts with 0, the first value
	// defined here will be the default
	Undefined PodCategory = iota
	Progressing
	Watching
	Converged
)
