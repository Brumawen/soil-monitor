package main

import (
	"flag"
	"log"
	"time"

	gopifinder "github.com/brumawen/gopi-finder/src"
	"github.com/brumawen/gopi-tools/src"
	"github.com/gorilla/mux"
	"github.com/onatm/clockwerk"
)

func main() {
	host := flag.String("h", "", "Host Name or IP Address.  (default All)")
	port := flag.Int("p", 20510, "Port Number to listen on.")
	verbose := flag.Bool("v", false, "Verbose logging.")
	timeout := flag.Int("t", 2, "Timeout in seconds to wait for a response from a IP probe.")
	mins := flag.Int("m", 5, "Number of minutes between measurements.")

	flag.Parse()

	// Create a new server
	s := Server{
		Host:           *host,
		PortNo:         *port,
		VerboseLogging: *verbose,
		Timeout:        *timeout,
		Router:         mux.NewRouter().StrictSlash(true),
		Monitor:        SoilMonitor{VerboseLogging: *verbose},
		Led:            gopitools.Led{GpioLed: 18, TurnOffOnClose: true},
		Finder:         gopifinder.Finder{VerboseLog: *verbose},
	}

	// Set the LED
	if err := s.Led.On(); err != nil {
		log.Println("Failed to switch on the LED.")
	}
	defer s.Led.Close()

	// Add the controllers
	s.AddController(new(MeasureController))

	// Run the monitor code to extract the values immedietely
	go s.Monitor.Run()

	// Start the scheduler
	cw := clockwerk.New()
	cw.Every(time.Duration(*mins) * time.Minute).Do(&s.Monitor)
	cw.Start()

	// Start the server
	err := s.ListenAndServe()
	if err != nil {
		log.Println(err)
	}
}
