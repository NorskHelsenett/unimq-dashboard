package models

import "time"

type Sample struct {
	T float64 `json:"t"`
	V float64 `json:"v"`
}

type RangeOptions struct {
	Vhost string
	Queue string
	Since time.Duration
	Step  time.Duration
}
