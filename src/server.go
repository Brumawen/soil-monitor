package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/brumawen/gopi-finder/src"
	"github.com/brumawen/gopi-tools/src"
	"github.com/gorilla/mux"
)

// Server defines the Web Server.
type Server struct {
	Host           string
	PortNo         int
	VerboseLogging bool
	Timeout        int
	Router         *mux.Router
	Finder         gopifinder.Finder
	Monitor        SoilMonitor
	Led            gopitools.Led
	Device         gopifinder.DeviceInfo
}

// AddController adds the specified web service controller to the Router
func (s *Server) AddController(c Controller) {
	c.AddController(s.Router, s)
}

// ListenAndServe starts the server
func (s *Server) ListenAndServe() error {
	// Register service with the Finder server
	go s.registerService()
	// Start the web service
	log.Println("Server listening on port", s.PortNo)
	return http.ListenAndServe(fmt.Sprintf("%v:%d", s.Host, s.PortNo), s.Router)
}

func (s *Server) registerService() {
	// Flash the LED
	s.Led.Flash(250)

	isReg := false

	for !isReg {
		_, err := s.Finder.FindDevices()
		if err != nil {
			log.Println("Error getting list of devices.")
		} else {
			isReg = true
		}
	}

	// Set the LED to solid
	s.Led.On()
}
