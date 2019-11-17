package main

import (
	"errors"
	"fmt"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

// Mqtt publishes the telemetry to a MQTT broker
type Mqtt struct {
	Srv               *Server     // Server instance
	LastUpdateAttempt time.Time   // Last time an update was attempted
	LastUpdate        time.Time   // Last time an update was published
	client            MQTT.Client // MQTT client

}

// Initialize starts up the MQTT client
func (m *Mqtt) Initialize() error {
	if !m.Srv.Config.EnableMqtt {
		m.logInfo("MQTT has been disabled")
		return nil
	}
	if m.Srv.Config.MqttHost == "" {
		m.logError("MQTT Host has not been configured.")
		m.Srv.Config.EnableMqtt = false
		return errors.New("host has not been configured")
	}
	if m.Srv.Config.MqttUsername == "" {
		m.logError("MQTT Username has not been configured.")
		m.Srv.Config.EnableMqtt = false
		return errors.New("username has not been configured")
	}
	if m.Srv.Config.MqttPassword == "" {
		m.logError("MQTT Password has not been configured.")
		m.Srv.Config.EnableMqtt = false
		return errors.New("password has not been configured")
	}

	// Connect and send meta information
	m.logInfo("Connecting to the MQTT Broker.")

	opts := MQTT.NewClientOptions()
	opts.AddBroker(m.Srv.Config.MqttHost)
	opts.SetUsername(m.Srv.Config.MqttUsername)
	opts.SetPassword(m.Srv.Config.MqttPassword)

	opts.SetConnectionLostHandler(func(client MQTT.Client, err error) {
		m.logError("Disconnected from MQTT Broker.", err.Error())
	})
	opts.SetOnConnectHandler(func(client MQTT.Client) {
		m.logInfo("Connected to the MQTT Broker.")
	})

	m.client = MQTT.NewClient(opts)
	if token := m.client.Connect(); token.Wait() && token.Error() != nil {
		m.logError("Error connecting to MQTT Broker.", token.Error())
		return token.Error()
	}

	return nil
}

// Close closes the MQTT client and disconnects
func (m *Mqtt) Close() {
	m.client.Disconnect(250)
}

// SendTelemetry sends the current states of the devices to the MQTT Broker
func (m *Mqtt) SendTelemetry(v Measurement) error {
	if !m.Srv.Config.EnableMqtt {
		return nil
	}

	if !v.Success {
		m.logInfo("Cannot publish values to MQTT.  There was an error with the values.")
		return nil
	}

	m.logInfo("Publishing telemetry to MQTT")
	m.LastUpdateAttempt = time.Now()

	if !m.client.IsConnected() {
		m.logInfo("Reconnecting to MQTT broker")
		if token := m.client.Connect(); token.Wait() && token.Error() != nil {
			m.logError("Error connecting to MQTT Broker.", token.Error())
			return token.Error()
		}
	}

	// Temperature
	m.logInfo("Publishing air temperature - ", fmt.Sprintf("%.1f", v.AirTemp), "C")
	token := m.client.Publish("home/garden/airtemp", byte(0), true, fmt.Sprintf("%.1f", v.AirTemp))
	if token.Wait() && token.Error() != nil {
		m.logError("Error sending air temperature state to MQTT Broker.", token.Error())
		return token.Error()
	}

	m.logInfo("Publishing soil temperature - ", fmt.Sprintf("%.1f", v.SoilTemp), "C")
	token = m.client.Publish("home/garden/soiltemp", byte(0), true, fmt.Sprintf("%.1f", v.SoilTemp))
	if token.Wait() && token.Error() != nil {
		m.logError("Error sending soil temperature state to MQTT Broker.", token.Error())
		return token.Error()
	}

	// Light
	m.logInfo("Publishing light - ", fmt.Sprintf("%.1f", v.Light), "%")
	token = m.client.Publish("home/garden/light", byte(0), true, fmt.Sprintf("%.1f", v.Light))
	if token.Wait() && token.Error() != nil {
		m.logError("Error sending light state to MQTT Broker.", token.Error())
		return token.Error()
	}
	// Moisture
	m.logInfo("Publishing moisture - ", fmt.Sprintf("%.1f", v.Moisture), "%")
	token = m.client.Publish("home/garden/moisture", byte(0), true, fmt.Sprintf("%.1f", v.Moisture))
	if token.Wait() && token.Error() != nil {
		m.logError("Error sending moisture state to MQTT Broker.", token.Error())
		return token.Error()
	}

	m.LastUpdate = time.Now()

	return nil
}

// logInfo logs an information message to the logger
func (m *Mqtt) logInfo(v ...interface{}) {
	a := fmt.Sprint(v...)
	logger.Info("Mqtt: [Inf] ", a)
}

// logError logs an error message to the logger
func (m *Mqtt) logError(v ...interface{}) {
	a := fmt.Sprint(v...)
	logger.Error("Mqtt [Err] ", a)
}
