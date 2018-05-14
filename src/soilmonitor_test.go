package main

import (
	"testing"
)

func TestCanMeasureValues(t *testing.T) {
	m := SoilMonitor{}
	v, err := m.MeasureValues()
	if err != nil {
		t.Error(err)
	}
	if v.Temperature == 0 {
		t.Error("Temperature is 0.")
	}
}
