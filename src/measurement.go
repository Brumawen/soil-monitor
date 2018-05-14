package main

import (
	"time"
)

type Measurement struct {
	Temperature  float64
	Light        float64
	Moisture     float64
	Success      bool
	Error        string
	DateMeasured time.Time
}
