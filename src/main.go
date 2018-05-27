package main

import (
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/kardianos/service"
)

var logger service.Logger

func main() {
	port := flag.Int("p", 20510, "Port Number to listen on.")
	timeout := flag.Int("t", 2, "Timeout in seconds to wait for a response from a IP probe.")
	mins := flag.Int("m", 5, "Number of minutes between measurements.")
	svcFlag := flag.String("service", "", "Service action.  Valid actions are: 'start', 'stop', 'restart', 'instal' and 'uninstall'")
	flag.Parse()

	// Create a new server
	s := &Server{
		PortNo:    *port,
		Timeout:   *timeout,
		SchedTime: *mins,
	}

	// Create the service
	svcConfig := &service.Config{
		Name:        "SoilMonitor",
		DisplayName: "Soil Monitor",
		Description: "Monitors soil temperature, humidity and ambient light.",
	}
	v, err := service.New(s, svcConfig)
	if err != nil {
		log.Fatal(err)
	}

	// Set up the logger
	errs := make(chan error, 5)
	logger, err = v.Logger(errs)
	if err != nil {
		log.Fatal(err)
	}
	go func() {
		for {
			err := <-errs
			if err != nil {
				log.Print(err)
			}
		}
	}()

	// Start the service
	if *svcFlag != "" {
		// Service control request
		if err := service.Control(v, *svcFlag); err != nil {
			e := err.Error()
			if strings.Contains(e, "Unknown action") {
				fmt.Println(*svcFlag, "is an invalid action.")
				fmt.Println("Valid actions are", service.ControlAction)
			} else {
				fmt.Println(err.Error())
			}
		}
	} else {
		// Start the service in debug if we are running in a terminal
		s.VerboseLogging = service.Interactive()
		if err := v.Run(); err != nil {
			log.Fatal(err)
		}
	}
}
