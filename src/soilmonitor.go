package main

import (
	"errors"
	"fmt"
	"net/http"
	"time"

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
		if m.Srv.Config.EnableThingspeak {
			// Send the measurement to Thingspeak
			m.logDebug("Sending result to Thingspeak.")
			err = m.sendToThingspeak(v)
			if err != nil {
				m.logError("Error sending result to Thingspeak. " + err.Error())
				v.Error = err.Error()
			}
		}
		// Append the measurement to the list
		m.Measurements = append(m.Measurements, v)
	}
	// Only keep the last 10 measurements
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

	// Get the temperature probe
	tmp := gopitools.OneWireTemp{}
	defer tmp.Close()
	tmp.ID = ""

	// Get the available one-wire devices
	okToRead := true
	m.logDebug("Getting one-wire device list.")
	devlst, err := gopitools.GetDeviceList()
	if err != nil {
		msg := "Error getting one-wire device list. " + err.Error() + "."
		m.logError(msg)
		errLst = append(errLst, msg)
	} else {
		if len(devlst) == 0 {
			m.Srv.LCD.SetItem("TEMP", "Temp", "No Cable")
			msg := "No temperature device found. Cable could be disconnected."
			m.logError(msg)
			errLst = append(errLst, msg)
			okToRead = false
		} else {
			m.logDebug("Reading temperature from ", devlst[0].Name)
			tmp.ID = devlst[0].ID
			temp, err := tmp.ReadTemp()
			if err != nil {
				m.Srv.LCD.SetItem("TEMP", "Temp", "Err")
				msg := "Error reading temperature. " + err.Error() + "."
				m.logError(msg)
				errLst = append(errLst, msg)
			} else {
				m.Srv.LCD.SetItem("TEMP", "Temp", fmt.Sprintf("%f", temp))
				v.Temperature = temp
			}
		}
	}

	if okToRead {
		// Read ambient light and moisture content
		m.logDebug("Reading Light and Moisture values")
		mcp := gopitools.Mcp3008{}
		defer mcp.Close()
		mcpVals, err := mcp.Read()
		if err != nil {
			m.Srv.LCD.SetItem("LIGHT", "Light", "Err")
			m.Srv.LCD.SetItem("MOISTURE", "Moisture", "Err")
			msg := "Error reading MCP3008 values. " + err.Error() + "."
			m.logError(msg)
			errLst = append(errLst, msg)
		} else {
			v.Light = 100 - mcpVals[0]
			m.Srv.LCD.SetItem("LIGHT", "Light", fmt.Sprintf("%f", v.Light))
			v.Moisture = mcpVals[1]
			m.Srv.LCD.SetItem("MOISTURE", "Moisture", fmt.Sprintf("%f", v.Moisture))
		}
	} else {
		m.Srv.LCD.SetItem("LIGHT", "Light", "No Cable")
		m.Srv.LCD.SetItem("MOISTURE", "Moisture", "No Cable")
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
	url := fmt.Sprintf("https://api.thingspeak.com/update?api_key=%s&field1=%f&field2=%f&field3=%f", key, v.Temperature, v.Light, v.Moisture)
	_, err := client.Get(url)
	if err != nil {
		return err
	}
	return nil
}

func (m *SoilMonitor) logDebug(v ...interface{}) {
	if m.Srv.VerboseLogging {
		a := fmt.Sprint(v)
		logger.Info("SoilMonitor: ", a[1:len(a)-1])
	}
}

func (m *SoilMonitor) logInfo(v ...interface{}) {
	a := fmt.Sprint(v)
	logger.Info("SoilMonitor: ", a[1:len(a)-1])
}

func (m *SoilMonitor) logError(v ...interface{}) {
	a := fmt.Sprint(v)
	logger.Error("SoilMonitor: ", a[1:len(a)-1])
}
