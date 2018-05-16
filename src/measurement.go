package main

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

// Measurement holds the values read from the component probes.
type Measurement struct {
	Temperature  float64
	Light        float64
	Moisture     float64
	Success      bool
	Error        string
	DateMeasured time.Time
}

// ReadFrom reads the string from the reader and deserializes it into the entity values
func (m *Measurement) ReadFrom(r io.ReadCloser) error {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	if b != nil && len(b) != 0 {
		if err := json.Unmarshal(b, &m); err != nil {
			return err
		}
	}
	return nil
}

// WriteTo serializes the entity and writes it to the http response
func (m *Measurement) WriteTo(w http.ResponseWriter) error {
	b, err := json.Marshal(m)
	if err != nil {
		return err
	}
	w.Header().Set("content-type", "application/json")
	w.Write(b)
	return nil
}

// Serialize serializes the entity and returns the serialized string
func (m *Measurement) Serialize() (string, error) {
	b, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// Deserialize deserializes the specified string into the entity values
func (m *Measurement) Deserialize(v string) error {
	err := json.Unmarshal([]byte(v), &m)
	if err != nil {
		return err
	}
	return nil
}
