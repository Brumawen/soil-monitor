package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/brumawen/gopi-finder/src"

	gopitools "github.com/brumawen/gopi-tools/src"
)

// SoilMonitor manages the monitoring of the soil measurement components
// and provides the latest readings.
type SoilMonitor struct {
	Srv          *Server
	LastRead     time.Time
	Measurements []Measurement
	IsRunning    bool
}

// Run is called from the scheduler (ClockWerk). This function will get the latest measurements
// and send the measurements to Thingspeak
// It will also keep the last hour's worth of measurements in a list.
func (m *SoilMonitor) Run() {
	m.logDebug("Starting run.")
	// Get the current measurements
	v, err := m.MeasureValues()
	if err != nil {
		m.logError("Error getting measurements. ", err.Error())
		m.Measurements = append(m.Measurements, Measurement{
			Success:      false,
			Error:        err.Error(),
			DateMeasured: time.Now(),
		})
	} else {
		// Send the measurement to Thingspeak
		m.logDebug("Sending result to Thingspeak.")
		err = m.sendToThingspeak(v)
		if err != nil {
			m.logError("Error sending result to Thingspeak. " + err.Error())
			v.Error = err.Error()
		}
		// Append the measurement to the list
		m.Measurements = append(m.Measurements, v)
	}
	// Only keep the last 10 measurements
	if len(m.Measurements) > 12 {
		// Remove the first item
		m.Measurements = m.Measurements[1:]
	}
	m.logDebug("Completed run.")
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
		m.logError("Error turning on power.", err.Error())
		return v, errors.New("Error turning on power. " + err.Error())
	}

	// wait 2 secs to let everthing stabilize
	time.Sleep(2 * time.Second)

	// Get the temperature probe
	tmp := gopitools.OneWireTemp{}
	defer tmp.Close()
	tmp.ID = ""

	// Get the available one-wire devices
	m.logDebug("Getting one-wire device list.")
	devlst, err := gopitools.GetDeviceList()
	if err != nil {
		m.logError("Error getting one-wire device list.", err.Error())
		return v, errors.New("Error getting one-wire device list." + err.Error())
	}
	if len(devlst) == 0 {
		m.logError("No temperature device found.")
		return v, errors.New("No temperature device found")
	}
	m.logDebug("Reading temperature.")
	tmp.ID = devlst[0].ID
	temp, err := tmp.ReadTemp()
	if err != nil {
		m.logError("Error reading temperature.", err.Error())
		return v, errors.New("Error reading temperature. " + err.Error())
	}
	v.Temperature = temp

	// Read ambient light and moisture content
	m.logDebug("Reading Light and Moisture values")
	mcp := gopitools.Mcp3008{}
	defer mcp.Close()
	mcpVals, err := mcp.Read()
	if err != nil {
		m.logError("Error reading MCP3008 values.", err.Error())
		return v, errors.New("Error reading MCP3008 values. " + err.Error())
	}
	v.Light = 100 - mcpVals[0]
	v.Moisture = mcpVals[1]

	// Switch off the power to the soil components
	m.logDebug("Turning off power")
	err = pwr.Off()
	if err != nil {
		m.logError("Error turning off power.", err.Error())
		v.Error = "Error turning off power. " + err.Error()
	}

	v.Success = true
	return v, nil
}

func (m *SoilMonitor) setStopped() {
	m.IsRunning = false
}

func (m *SoilMonitor) sendToThingspeak(v Measurement) error {
	if _, err := os.Stat("ts-api-key"); os.IsNotExist(err) {
		// Thingspeak API key file is missing
		return errors.New("API key file 'ts-api-key' is missing")
	}

	// Get the thingspeak api key
	key, err := gopifinder.ReadAllText("ts-api-key")
	if err != nil {
		return errors.New("Error reading API Key from 'ts-api-key' file. " + err.Error())
	}
	client := http.Client{}
	url := fmt.Sprintf("https://api.thingspeak.com/update?api_key=%s&field1=%f&field2=%f&field3=%f", key, v.Temperature, v.Light, v.Moisture)
	_, err = client.Get(url)
	if err != nil {
		return err
	}
	return nil
}

func (m *SoilMonitor) logDebug(v ...interface{}) {
	if m.Srv.VerboseLogging {
		a := fmt.Sprint(v)
		logger.Info("Server: ", a[1:len(a)-1])
	}
}

func (m *SoilMonitor) logInfo(v ...interface{}) {
	a := fmt.Sprint(v)
	logger.Info("Server: ", a[1:len(a)-1])
}

func (m *SoilMonitor) logError(v ...interface{}) {
	a := fmt.Sprint(v)
	logger.Error("Server: ", a[1:len(a)-1])
}
