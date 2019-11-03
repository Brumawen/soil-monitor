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
	if v.SoilTemp == 0 {
		t.Error("Soil temperature is 0.")
	}
	if v.AirTemp == 0 {
		t.Error("Air temperature is 0")
	}
}
