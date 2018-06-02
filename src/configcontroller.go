package main

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// ConfigController handles the Web Methods for configuring the module.
type ConfigController struct {
	Srv *Server
}

// ConfigPageData holds the data used to write to the configuration page.
type ConfigPageData struct {
	Period           int
	EnableThingspeak string
	ThingspeakID     string
}

// AddController adds the controller routes to the router
func (c *ConfigController) AddController(router *mux.Router, s *Server) {
	c.Srv = s
	router.Path("/config.html").Handler(http.HandlerFunc(c.handleConfigWebPage))
	router.Methods("GET").Path("/config/get").Name("GetConfig").
		Handler(Logger(c, http.HandlerFunc(c.handleGetConfig)))
	router.Methods("POST").Path("/config/set").Name("SetConfig").
		Handler(Logger(c, http.HandlerFunc(c.handleSetConfig)))
}

func (c *ConfigController) handleConfigWebPage(w http.ResponseWriter, r *http.Request) {
	t := template.Must(template.ParseFiles("./html/config.html"))

	v := ConfigPageData{
		Period:       c.Srv.Config.Period,
		ThingspeakID: c.Srv.Config.ThingspeakID,
	}
	if c.Srv.Config.EnableThingspeak {
		v.EnableThingspeak = "checked"
	}

	t.Execute(w, v)
}

func (c *ConfigController) handleGetConfig(w http.ResponseWriter, r *http.Request) {
	if err := c.Srv.Config.WriteTo(w); err != nil {
		http.Error(w, "Error serializing configuration. "+err.Error(), 500)
	}
}

func (c *ConfigController) handleSetConfig(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	pd := r.Form.Get("period")
	en := r.Form.Get("enableTS")
	id := r.Form.Get("tsID")

	if en == "on" && id == "" {
		http.Error(w, "Thingspeak ID must be specified", 500)
		return
	}

	c.LogInfo("Setting new configuration values.")
	v, err := strconv.Atoi(pd)
	if err != nil {
		http.Error(w, "Failed to convert "+pd+" to an integer.", 500)
		return
	}

	c.Srv.Config.Period = v
	c.Srv.Config.EnableThingspeak = (en == "on")
	c.Srv.Config.ThingspeakID = id
	c.Srv.Config.WriteToFile("config.json")
}

// LogInfo is used to log information messages for this controller.
func (c *ConfigController) LogInfo(v ...interface{}) {
	a := fmt.Sprint(v)
	logger.Info("ConfigController: ", a[1:len(a)-1])
}

// LogError is used to log error messages for this controller.
func (c *ConfigController) LogError(v ...interface{}) {
	a := fmt.Sprint(v)
	logger.Error("ConfigController: ", a[1:len(a)-1])
}
