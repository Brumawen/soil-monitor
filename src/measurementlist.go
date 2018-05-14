package main

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
)

// MeasurementList holds a list of measurements
type MeasurementList struct {
	Measurements []Measurement `json:"measurements"`
}

// ReadFrom reads the string from the reader and deserializes it into the entity values
func (s *MeasurementList) ReadFrom(r io.ReadCloser) error {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	if b != nil && len(b) != 0 {
		if err := json.Unmarshal(b, &s); err != nil {
			return err
		}
	}
	return nil
}

// WriteTo serializes the entity and writes it to the http response
func (s *MeasurementList) WriteTo(w http.ResponseWriter) error {
	b, err := json.Marshal(s)
	if err != nil {
		return err
	}
	w.Header().Set("content-type", "application/json")
	w.Write(b)
	return nil
}

// Serialize serializes the entity and returns the serialized string
func (s *MeasurementList) Serialize() (string, error) {
	b, err := json.Marshal(s)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// Deserialize deserializes the specified string into the entity values
func (s *MeasurementList) Deserialize(v string) error {
	err := json.Unmarshal([]byte(v), &s)
	if err != nil {
		return err
	}
	return nil
}
