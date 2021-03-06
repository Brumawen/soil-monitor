package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	gopifinder "github.com/brumawen/gopi-finder/src"
	gopitools "github.com/brumawen/gopi-tools/src"
	"github.com/gorilla/mux"
	"github.com/kardianos/service"
	"github.com/onatm/clockwerk"
)

// Server defines the Web Server.
type Server struct {
	PortNo         int                  // Port No the server will listen on
	VerboseLogging bool                 // Verbose logging on/ off
	Timeout        int                  // Timeout waiting for a response from an IP probe.  Defaults to 2 seconds.
	Config         *Config              // Configuration settings
	Finder         gopifinder.Finder    // Finder client - used to find other devices
	Monitor        SoilMonitor          // Soil monitor module
	MqttClient     *Mqtt                // MQTT client
	LCD            *Display             // LCD display
	Led            gopitools.Led        // LED module
	exit           chan struct{}        // Exit flag
	shutdown       chan struct{}        // Shutdown complete flag
	http           *http.Server         // HTTP server
	router         *mux.Router          // HTTP router
	cw             *clockwerk.Clockwerk // Clockwerk scheduler
	isregistering  bool                 // Indicates that a registration is currently ongoing
}

// Start is called when the service is starting
func (s *Server) Start(v service.Service) error {
	s.logInfo("Service starting")

	// Make sure the working directory is the same as the application exe
	ap, err := os.Executable()
	if err != nil {
		s.logError("Error getting the executable path.", err.Error())
	} else {
		wd, err := os.Getwd()
		if err != nil {
			s.logError("Error getting current working directory.", err.Error())
		} else {
			ad := filepath.Dir(ap)
			s.logInfo("Current application path is", ad)
			if ad != wd {
				if err := os.Chdir(ad); err != nil {
					s.logError("Error chaning working directory.", err.Error())
				}
			}
		}
	}

	// Create a channel that will be used to block until the Stop signal is received
	s.exit = make(chan struct{})
	go s.run()
	return nil
}

// Stop is called when the service is stopping
func (s *Server) Stop(v service.Service) error {
	s.logInfo("Service stopping")
	// Close the channel, this will automatically release the block
	s.shutdown = make(chan struct{})
	close(s.exit)
	// Wait for the shutdown to complete
	_ = <-s.shutdown
	return nil
}

// run will start up and run the service and wait for a Stop signal
func (s *Server) run() {
	if s.PortNo < 0 {
		s.PortNo = 20510
	}
	s.Monitor.Srv = s
	s.Finder.Logger = logger
	s.Finder.VerboseLogging = service.Interactive()

	// Get the configuration
	if s.Config == nil {
		s.Config = &Config{}
	}
	s.Config.ReadFromFile("config.json")

	// Create a router
	s.router = mux.NewRouter().StrictSlash(true)
	s.router.PathPrefix("/assets/").Handler(http.StripPrefix("/assets/", http.FileServer(http.Dir("./html/assets"))))

	// Add the controllers
	s.addController(new(MeasureController))
	s.addController(new(LogController))
	s.addController(new(ConfigController))

	// Create an HTTP server
	s.http = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.PortNo),
		Handler: s.router,
	}

	// Set the LED
	s.Led = gopitools.Led{GpioLed: 18}
	if err := s.Led.On(); err != nil {
		s.logError("Failed to switch on the LED.", err.Error())
	}

	// Set the display
	s.LCD = &Display{ShowTime: 5}
	s.LCD.SetItem("IP", "No IP", "")
	s.LCD.SetItem("AIRTEMP", "AirTemp", "")
	s.LCD.SetItem("SOILTEMP", "SoilTemp", "")
	s.LCD.SetItem("LIGHT", "Light", "")
	s.LCD.SetItem("MOISTURE", "Moisture", "")
	s.LCD.Start()

	if s.MqttClient == nil {
		s.MqttClient = &Mqtt{}
		s.MqttClient.Srv = s
	}

	go func() {
		// Register service with the Finder server
		go s.RegisterService()

		s.MqttClient.Initialize()

		// Read the values immedietely
		s.Monitor.Run()

		// Start the scheduler
		s.StartSchedule()
	}()

	// Start the web server
	go func() {
		s.logInfo("Server listening on port", s.PortNo)
		if err := s.http.ListenAndServe(); err != nil {
			msg := err.Error()
			if !strings.Contains(msg, "http: Server closed") {
				s.logError("Error starting Web Server.", err.Error())
			}
		}
	}()

	// Wait for an exit signal
	_ = <-s.exit

	// Turn off the LED
	if err := s.Led.Off(); err != nil {
		s.logError("Failed to turn off the LED.", err.Error())
	}

	// Shutdown the HTTP server
	s.http.Shutdown(nil)

	s.LCD.Stop()

	s.logDebug("Shutdown complete")
	close(s.shutdown)
}

// StartSchedule will start up the schedule for measuring the values
func (s *Server) StartSchedule() {
	if s.Config.Period <= 0 {
		s.Config.Period = 5
	}
	if s.cw != nil {
		s.cw.Stop()
	}
	s.cw = clockwerk.New()
	s.cw.Every(time.Duration(s.Config.Period) * time.Minute).Do(&s.Monitor)
	s.cw.Start()
}

// AddController adds the specified web service controller to the Router
func (s *Server) addController(c Controller) {
	c.AddController(s.router, s)
}

// RegisterService will register the service with the devices on the network
func (s *Server) RegisterService() {
	if s.isregistering {
		return
	}
	s.isregistering = true
	isReg := false
	s.logDebug("Starting service registration.")
	for !isReg {
		s.logDebug("RegisterService: Getting device info")
		d, err := gopifinder.NewDeviceInfo()
		if err != nil {
			s.logError("Error getting device info.", err.Error())
		}
		s.logDebug("RegisterService: Creating service")
		sv := d.CreateService("SoilMonitor")
		sv.PortNo = s.PortNo

		if sv.IPAddress == "" {
			s.logDebug("RegisterService: No IP address found.")
			s.LCD.SetItem("IP", "No IP", "")
		} else {
			s.logDebug("RegisterService: Using IP address", sv.IPAddress)
			ipArr := strings.Split(sv.IPAddress, ".")
			if len(ipArr) == 4 {
				s.LCD.SetItem("IP", ipArr[0]+"."+ipArr[1]+".", ipArr[2]+"."+ipArr[3])
			}
		}

		s.logDebug("Reg: Finding devices")
		_, err = s.Finder.FindDevices()
		if err != nil {
			s.logError("RegisterService: Error getting list of devices.", err.Error())
		} else {
			if len(s.Finder.Devices) == 0 {
				s.logDebug("RegisterService: Sleeping")
				time.Sleep(15 * time.Second)
			} else {
				// Register the services with the devices
				s.logDebug("RegisterService: Registering the service.")
				s.Finder.RegisterServices([]gopifinder.ServiceInfo{sv})
				isReg = true
			}
		}
	}
	s.logDebug("Completed service registration.")
	s.isregistering = false
}

// logDebug logs a debug message to the logger
func (s *Server) logDebug(v ...interface{}) {
	if s.VerboseLogging {
		a := fmt.Sprint(v...)
		logger.Info("Server: ", a)
	}
}

// logInfo logs an information message to the logger
func (s *Server) logInfo(v ...interface{}) {
	a := fmt.Sprint(v...)
	logger.Info("Server: ", a)
}

// logError logs an error message to the logger
func (s *Server) logError(v ...interface{}) {
	a := fmt.Sprint(v...)
	logger.Error("Server: ", a)
}
