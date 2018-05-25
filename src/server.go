package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/brumawen/gopi-finder/src"
	"github.com/brumawen/gopi-tools/src"
	"github.com/gorilla/mux"
	"github.com/kardianos/service"
)

// Server defines the Web Server.
type Server struct {
	PortNo         int                   // Port No the server will listen on
	VerboseLogging bool                  // Verbose logging on/ off
	Timeout        int                   // Timeout waiting for a response from an IP probe.  Defaults to 2 seconds.
	SchedTime 	   int					 // Schedule time (mins)
	finder         gopifinder.Finder     // Finder client - used to find other devices
	monitor        *SoilMonitor          // Soil monitor module
	led            gopitools.Led         // LED module
	device         gopifinder.DeviceInfo // This server's device
	exit           chan struct{}         // Exit flag
	http           *http.Server          // HTTP server
	router         *mux.Router           // HTTP router
	cw			   *clockwerk.Clockwerk  // Clockwerk scheduler
}

// Start is called when the service is starting
func (s *Server) Start(v service.Service) error {
	s.logInfo("Service starting")
	// Create a channel that will be used to block until the Stop signal is received
	s.exit = make(chan struct{})
	go s.run()
	return nil
}

// Stop is called when the service is stopping
func (s *Server) Stop(v service.Service) error {
	s.logInfo("Service stopping")
	// Close the channel, this will automatically release the block
	close(s.exit)
	return nil
}

// run will start up and run the service and wait for a Stop signal
func (s *Server) run() {
	if s.PortNo < 0 {
		s.PortNo = 20510
	}
	s.logInfo("Server listening on port", s.PortNo)

	// Create a router
	s.router = mux.NewRouter().StrictSlash(true)

	// Add the controllers
	s.addController(new(MeasureController))

	// Create an HTTP server
	s.http = &http.Server{
		Addr: fmt.Sprintf(":%d", s.PortNo)
		Handler: s.router
	}

	// Set the LED
	s.led := gopitools.Led{GpioLed: 18, TurnOffOnClose: true}
	if err := s.led.On(); err != nil {
		s.logError("Failed to switch on the LED.", err.Error())
	}
	defer s.Led.Close()

	go func() {
		
	}()

	// Start the scheduler
	go func() {
		// Create the soil monitor object
		s.monitor = SoilMonitor{
			VerboseLogging: s.VerboseLogging,
			Srv: s,
		}

		//
		s.finder = gopifinder.Finder{VerboseLog: s.VerboseLogging}

		// Read the values immedietely
		s.monitor.Run()

		// Start the scheduler
		cw := clockwerk.New()
		cw.Every(time.Duration(s.SchedTime) * time.Minute).Do(s.monitor)
		cw.Start()
	}()

	// Start the web server
	go func() {
		// Register service with the Finder server
		go s.registerService()

		if err := s.http.ListenAndServe(); err != nil {
			s.logError("Error starting Web Server.", err.Error())
		}
	}()

	// Wait for an exit signal
	_ = <-s.exit

	// Turn off the LED
	if err := s.led.Off(); err != nil {
		s.logError("Failed to turn off the LED.", err.Error())
	}

	// Shutdown the HTTP server
	s.http.Shutdown(nil)
}

// AddController adds the specified web service controller to the Router
func (s *Server) addController(c Controller) {
	c.AddController(s.Router, s)
}

func (s *Server) registerService() {
	// Flash the LED
	s.Led.Flash(250)

	isReg := false

	for !isReg {
		_, err := s.Finder.FindDevices()
		if err != nil {
			s.logError("Error getting list of devices.", err.Error())
		} else {
			if len(s.Finder.Devices) == 0 {
				time.Sleep(15 * time.Second)
			} else {
				// Register the services with the devices
				s.Finder.RegisterServices(sl)
				isReg = true
			}
		}
	}

	// Set the LED to solid
	s.Led.On()
}

// logDebug logs a debug message to the logger
func (s *Server) logDebug(v ...interface{}) {
	if s.VerboseLogging {
		a := fmt.Sprint(v)
		logger.Info("Server: ", a[1:len(a)-1])
	}
}

// logInfo logs an information message to the logger
func (s *Server) logInfo(v ...interface{}) {
	a := fmt.Sprint(v)
	logger.Info("Server: ", a[1:len(a)-1])
}

// logError logs an error message to the logger
func (s *Server) logError(v ...interface{}) {
	a := fmt.Sprint(v)
	logger.Error("Server: ", a[1:len(a)-1])
}
