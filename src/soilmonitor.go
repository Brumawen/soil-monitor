package main

import (
	"errors"
	"fmt"
	"math"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"time"

	gopitools "github.com/brumawen/gopi-tools/src"
)

// SoilMonitor manages the monitoring of the soil measurement components
// and provides the latest readings.
type SoilMonitor struct {
	Srv             *Server       // Server instance
	LastRead        time.Time     // Last time the measurement was taken
	Measurements    []Measurement // Last 10 measurements
	LastMeasurement Measurement   // Last successful measurement
	IsRunning       bool          // Is the monitor running
}

// Run is called from the scheduler (ClockWerk). This function will get the latest measurements
// and send the measurements to Thingspeak
// It will also keep the last 12 measurements in a list.
func (m *SoilMonitor) Run() {
	// Rerun a registration
	go m.Srv.RegisterService()

	m.logDebug("Starting measurement run.")
	// Get the current measurements
	v, err := m.MeasureValues()
	if err != nil {
		m.Measurements = append(m.Measurements, Measurement{
			Success:      false,
			Error:        err.Error(),
			DateMeasured: time.Now(),
		})
	} else {
		// Thingspeak
		if m.Srv.Config.EnableThingspeak {
			// Send the measurement to Thingspeak
			m.logDebug("Sending result to Thingspeak.")
			err = m.sendToThingspeak(v)
			if err != nil {
				m.logError("Error sending result to Thingspeak. " + err.Error())
				v.Error = err.Error()
			}
		}
		// MQTT
		if m.Srv.Config.EnableMqtt {
			// Send the measurement to MQTT broker
			m.logDebug("Sending result to MQTT.")
			err := m.Srv.MqttClient.SendTelemetry(v)
			if err != nil {
				m.logError("Error sending result to MQTT broker. " + err.Error())
				v.Error = err.Error()
			}
		}

		// Append the measurement to the list
		m.Measurements = append(m.Measurements, v)
	}
	// Only keep the last 12 measurements
	if len(m.Measurements) > 12 {
		// Remove the first item
		m.Measurements = m.Measurements[1:]
	}
	m.logDebug("Completed measurement run.")
}

// MeasureValues will measure the values from the component probes.
func (m *SoilMonitor) MeasureValues() (Measurement, error) {
	if m.IsRunning {
		if len(m.Measurements) == 0 {
			return Measurement{}, nil
		}
		return m.Measurements[len(m.Measurements)-1], nil
	}
	m.IsRunning = true
	defer m.setStopped()

	m.logDebug("Reading measurements.")

	v := Measurement{
		DateMeasured: time.Now(),
	}

	// Switch on the power to the soil components
	m.logDebug("Turning on power.")
	pwr := gopitools.Pin{GpioNo: 22, TurnOffOnClose: true}
	defer pwr.Close()
	if err := pwr.On(); err != nil {
		msg := "Error turning on power. " + err.Error() + "."
		m.logError(msg)
		return v, errors.New(msg)
	}

	// wait 2 secs to let everthing stabilize
	time.Sleep(2 * time.Second)

	errLst := []string{}

	// Get the available one-wire devices
	m.logDebug("Getting one-wire device list.")
	devlst, err := gopitools.GetDeviceList()
	if err != nil {
		msg := "Error getting one-wire device list. " + err.Error() + "."
		m.logError(msg)
		errLst = append(errLst, msg)
	}

	// Read the Air Temperature
	airTemp := gopitools.OneWireTemp{}
	defer airTemp.Close()
	airTemp.ID = m.Srv.Config.AirTempID
	if !airTemp.IsInDevices(devlst) {
		m.Srv.LCD.SetItem("AIRTEMP", "AirTemp", "No Cable")
		msg := "No air temperature device '" + airTemp.ID + "' found. Cable could be disconnected."
		m.logError(msg)
	} else {
		m.logDebug("Reading air temperature from ", airTemp.ID)
		temp, err := airTemp.ReadTemp()
		if err != nil {
			m.Srv.LCD.SetItem("AIRTEMP", "AirTemp", "Err")
			msg := "Error reading air temperature. " + err.Error() + "."
			m.logError(msg)
			errLst = append(errLst, msg)
		} else {
			m.Srv.LCD.SetItem("AIRTEMP", "AirTemp", fmt.Sprintf("%f", temp))
			v.AirTemp = temp
		}
	}

	// Read the soil temperature
	soilTemp := gopitools.OneWireTemp{}
	defer soilTemp.Close()
	soilTemp.ID = m.Srv.Config.SoilTempID
	if !airTemp.IsInDevices(devlst) {
		m.Srv.LCD.SetItem("SOILTEMP", "SoilTemp", "No Cable")
		msg := "No soil temperature device '" + airTemp.ID + "' found. Cable could be disconnected."
		m.logError(msg)
	} else {
		m.logDebug("Reading soil temperature from ", airTemp.ID)
		temp, err := soilTemp.ReadTemp()
		if err != nil {
			m.Srv.LCD.SetItem("SOILTEMP", "SoilTemp", "Err")
			msg := "Error reading soil temperature. " + err.Error() + "."
			m.logError(msg)
			errLst = append(errLst, msg)
		} else {
			m.Srv.LCD.SetItem("SOILTEMP", "SoilTemp", fmt.Sprintf("%f", temp))
			v.SoilTemp = temp
		}
	}

	// Read ambient light and moisture content
	m.logDebug("Reading Light and Moisture values")
	out, err := exec.Command("python", "mcp3008.py").CombinedOutput()
	if err != nil {
		msg := "Failed to get light and moisture content values. " + err.Error() + "."
		m.logError(msg)
		errLst = append(errLst, msg)
		m.Srv.LCD.SetItem("LIGHT", "Light", "Err")
		m.Srv.LCD.SetItem("MOISTURE", "Moisture", "Err")
	} else {
		outStr := strings.TrimSpace(string(out))
		m.logDebug("Values returned =", outStr)
		mcpVals := strings.Split(outStr, ",")

		if f, err := strconv.ParseFloat(mcpVals[0], 64); err != nil {
			msg := "Failed to get light value. " + err.Error() + "."
			m.logError(msg)
			errLst = append(errLst, msg)
			m.Srv.LCD.SetItem("LIGHT", "Light", "Err")
		} else {
			v.Light = math.Round(((100 - (f * 100)) * 100) / 100)
			m.Srv.LCD.SetItem("LIGHT", "Light", fmt.Sprintf("%f", v.Light))
		}

		if f, err := strconv.ParseFloat(mcpVals[1], 64); err != nil {
			msg := "Failed to get moisture content value. " + err.Error() + "."
			m.logError(msg)
			errLst = append(errLst, msg)
			m.Srv.LCD.SetItem("MOISTURE", "Moisture", "Err")
		} else {
			v.Moisture = math.Round((f * 100) / 100)
			m.Srv.LCD.SetItem("MOISTURE", "Moisture", fmt.Sprintf("%f", v.Moisture))
		}
	}

	// Switch off the power to the soil components
	m.logDebug("Turning off power")
	err = pwr.Off()
	if err != nil {
		msg := "Error turning off power. " + err.Error() + "."
		m.logError(msg)
		errLst = append(errLst, msg)
	}

	if len(errLst) == 0 {
		v.Success = true
		return v, nil
	}

	msg := ""
	for _, i := range errLst {
		if msg != "" {
			msg = msg + "\n"
		}
		msg = msg + i
	}
	return v, errors.New(msg)
}

func (m *SoilMonitor) setStopped() {
	m.IsRunning = false
}

func (m *SoilMonitor) sendToThingspeak(v Measurement) error {
	key := m.Srv.Config.ThingspeakID
	if key == "" {
		return errors.New("Thingspeak API ID has not been configured")
	}

	client := http.Client{}
	url := fmt.Sprintf("https://api.thingspeak.com/update?api_key=%s&field1=%f&field2=%f&field3=%f", key, v.SoilTemp, v.Light, v.Moisture)
	_, err := client.Get(url)
	if err != nil {
		return err
	}
	return nil
}

func (m *SoilMonitor) logDebug(v ...interface{}) {
	if m.Srv.VerboseLogging {
		a := fmt.Sprint(v...)
		logger.Info("SoilMonitor: ", a)
	}
}

func (m *SoilMonitor) logInfo(v ...interface{}) {
	a := fmt.Sprint(v...)
	logger.Info("SoilMonitor: ", a)
}

func (m *SoilMonitor) logError(v ...interface{}) {
	a := fmt.Sprint(v...)
	logger.Error("SoilMonitor: ", a)
}
