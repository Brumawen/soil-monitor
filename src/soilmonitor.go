package main

import (
	"errors"
	"log"
	"time"

	gopitools "github.com/brumawen/gopi-tools/src"
)

// SoilMonitor manages the monitoring of the soil measurement components
// and provides the latest readings.
type SoilMonitor struct {
	LastRead       time.Time
	Measurements   []Measurement
	VerboseLogging bool
}

func (m *SoilMonitor) Run() {
	// Get the current device information
	

	v, err := m.MeasureValues()
	if err != nil {
		m.Measurements = append(m.Measurements, Measurement{
			Success:      false,
			Error:        err.Error(),
			DateMeasured: time.Now(),
		})
	} else {
		m.Measurements = append(m.Measurements, v)
	}
	// Only keep the last 10 measurements
	if len(m.Measurements) > 12 {
		// Remove the first item
		m.Measurements = m.Measurements[1:]
	}
}

func (m *SoilMonitor) MeasureValues() (Measurement, error) {
	if m.VerboseLogging {
		log.Println("Measuring values.")
	}

	v := Measurement{
		DateMeasured: time.Now(),
	}

	// Switch on the power to the soil components
	if m.VerboseLogging {
		log.Println("Turning on power.")
	}
	pwr := gopitools.Pin{GpioNo: 22, TurnOffOnClose: true}
	defer pwr.Close()
	if err := pwr.On(); err != nil {
		return v, errors.New("Error turning on power. " + err.Error())
	}

	// wait 2 secs to let everthing stabilize
	time.Sleep(2 * time.Second)

	// Get the temperature probe
	tmp := gopitools.OneWireTemp{}
	defer tmp.Close()
	tmp.ID = ""

	// Get the available one-wire devices
	if m.VerboseLogging {
		log.Println("Getting one-wire device list.")
	}
	devlst, err := gopitools.GetDeviceList()
	if err != nil {
		return v, errors.New("Error getting temperature device list." + err.Error())
	}
	if len(devlst) == 0 {
		return v, errors.New("No temperature device found.")
	}
	if m.VerboseLogging {
		log.Println("Reading temperature.")
	}
	tmp.ID = devlst[0].ID
	temp, err := tmp.ReadTemp()
	if err != nil {
		return v, errors.New("Error reading temperature. " + err.Error())
	}
	v.Temperature = temp

	// Read ambient light and moisture content
	if m.VerboseLogging {
		log.Println("Reading Light and Moisture values")
	}
	mcp := gopitools.Mcp3008{}
	defer mcp.Close()
	mcpVals, err := mcp.Read()
	if err != nil {
		return v, errors.New("Error reading MCP3008 values. " + err.Error())
	}
	v.Light = 100 - mcpVals[0]
	v.Moisture = mcpVals[1]

	// Switch off the power to the soil components
	if m.VerboseLogging {
		log.Println("Turning off power")
	}
	err = pwr.Off()
	if err != nil {
		v.Error = "Error turning off power. " + err.Error()
	}

	v.Success = true
	return v, nil
}
