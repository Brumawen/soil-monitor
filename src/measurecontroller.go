package main

import "github.com/gorilla/mux"
import "net/http"

type MeasureController struct {
	Srv *Server
}

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
