package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

// MeasureController handles the Web Methods for measuring the values.
type MeasureController struct {
	Srv *Server
}

// AddController adds the controller routes to the router
func (c *MeasureController) AddController(router *mux.Router, s *Server) {
	c.Srv = s
	router.Methods("GET").Path("/measure/get").Name("GetMeasurements").
		Handler(Logger(http.HandlerFunc(c.handleGetMeasure)))
	router.Methods("GET").Path("/measure/getcurrent").Name("GetCurrent").
		Handler(Logger(http.HandlerFunc(c.handleGetCurrent)))
}

func (c *MeasureController) handleGetMeasure(w http.ResponseWriter, r *http.Request) {
	l := MeasurementList{
		Measurements: c.Srv.Monitor.Measurements,
	}
	if err := l.WriteTo(w); err != nil {
		http.Error(w, "Error serializing list. "+err.Error(), 500)
	}
}

func (c *MeasureController) handleGetCurrent(w http.ResponseWriter, r *http.Request) {
	if v, err := c.Srv.Monitor.MeasureValues(); err != nil {
		http.Error(w, "Error getting measurements. "+err.Error(), 500)
	} else {
		if err := v.WriteTo(w); err != nil {
			http.Error(w, "Error serializing measurements. "+err.Error(), 500)
		}
	}

}

// LogInfo is used to log information messages for this controller.
func (c *MeasureController) LogInfo(v ...interface{}) {
	a := fmt.Sprint(v)
	logger.Info("MeasureController: ", a[1:len(a)-1])
}
