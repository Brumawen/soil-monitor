package main

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"os"
)

// Config holds the configuration required for the Soil Monitor module.
type Config struct {
	Period           int    `json:"period"`           // The update period (in minutes)
	EnableThingspeak bool   `json:"enableThingspeak"` // Enable Thingspeak integration
	ThingspeakID     string `json:"thingspeakID"`     // Thingspeak ID
	EnableMqtt       bool   `json:"enableMqtt"`       // Enable MQTT integration
	MqttHost         string `json:"mqttHost"`         // MQTT Host
	MqttUsername     string `json:"mqttUsername"`     // MQTT Username
	MqttPassword     string `json:"mqttPassword"`     // MQTT password
	AirTempID        string `json:"airTempId"`        // ID of the Air temperature sensor
	SoilTempID       string `json:"soilTempId"`       // ID of the Soil temperature sensor
}

// ReadFromFile will read the configuration settings from the specified file
func (c *Config) ReadFromFile(path string) error {
	_, err := os.Stat(path)
	if !os.IsNotExist(err) {
		b, err := ioutil.ReadFile(path)
		if err == nil {
			err = json.Unmarshal(b, &c)
		}
	}
	c.setDefaults()
	return err
}

// WriteToFile will write the configuration settings to the specified file
func (c *Config) WriteToFile(path string) error {
	b, err := json.Marshal(c)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, b, 0666)
}

// ReadFrom reads the string from the reader and deserializes it into the entity values
func (c *Config) ReadFrom(r io.ReadCloser) error {
	b, err := ioutil.ReadAll(r)
	if err == nil {
		if b != nil && len(b) != 0 {
			err = json.Unmarshal(b, &c)
		}
	}
	c.setDefaults()
	return err
}

// WriteTo serializes the entity and writes it to the http response
func (c *Config) WriteTo(w http.ResponseWriter) error {
	b, err := json.Marshal(c)
	if err != nil {
		return err
	}
	w.Header().Set("content-type", "application/json")
	w.Write(b)
	return nil
}

// Serialize serializes the entity and returns the serialized string
func (c *Config) Serialize() (string, error) {
	b, err := json.Marshal(c)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// Deserialize deserializes the specified string into the entity values
func (c *Config) Deserialize(v string) error {
	err := json.Unmarshal([]byte(v), &c)
	c.setDefaults()
	return err
}

func (c *Config) setDefaults() {
	if c.Period <= 0 {
		c.Period = 5
	}
}
