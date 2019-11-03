package main

import "testing"

func TestMaskValue(t *testing.T) {
	c := ConfigController{}
	m := c.MaskValue()
	if len(m) != 20 {
		t.Error("Mask is not 20 characters long")
	}
}
