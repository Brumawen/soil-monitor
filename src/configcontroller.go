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
	EnableMqtt       string
	MqttHost         string
	MqttUsername     string
	MqttPassword     string
	AirTempID        string
	SoilTempID       string
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
		MqttHost:     c.Srv.Config.MqttHost,
		MqttUsername: c.Srv.Config.MqttUsername,
		MqttPassword: c.MaskValue(),
		AirTempID:    c.Srv.Config.AirTempID,
		SoilTempID:   c.Srv.Config.SoilTempID,
	}
	if c.Srv.Config.EnableThingspeak {
		v.EnableThingspeak = "checked"
	}
	if c.Srv.Config.EnableMqtt {
		v.EnableMqtt = "checked"
	}

	t.Execute(w, v)
}

// MaskValue returns a string with 20 stars
func (c *ConfigController) MaskValue() string {
	r := make([]rune, 20)
	for i := range r {
		r[i] = '*'
	}
	return string(r)
}

func (c *ConfigController) handleGetConfig(w http.ResponseWriter, r *http.Request) {
	if err := c.Srv.Config.WriteTo(w); err != nil {
		http.Error(w, "Error serializing configuration. "+err.Error(), 500)
	}
}

func (c *ConfigController) handleSetConfig(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	pd := r.Form.Get("period")

	ents := r.Form.Get("enableTS")
	tsid := r.Form.Get("tsID")

	enmq := r.Form.Get("enableMQTT")
	mhst := r.Form.Get("mqttHost")
	musr := r.Form.Get("mqttUser")
	mpwd := r.Form.Get("mqttPword")
	mask := c.MaskValue()

	aid := r.Form.Get("airTempID")
	sid := r.Form.Get("soilTempID")

	if ents == "on" && tsid == "" {
		http.Error(w, "Thingspeak ID must be specified", 500)
		return
	}

	if enmq == "on" && mhst == "" {
		http.Error(w, "MQTT Host must be specified", 500)
		return
	}

	c.LogInfo("Setting new configuration values.")
	v, err := strconv.Atoi(pd)
	if err != nil {
		http.Error(w, "Failed to convert "+pd+" to an integer.", 500)
		return
	}

	// Update the configuration values
	c.Srv.Config.Period = v

	c.Srv.Config.EnableThingspeak = (ents == "on")
	c.Srv.Config.ThingspeakID = tsid

	c.Srv.Config.EnableMqtt = (enmq == "on")
	c.Srv.Config.MqttHost = mhst
	c.Srv.Config.MqttUsername = musr
	if mpwd != mask {
		c.Srv.Config.MqttPassword = mpwd
	}

	c.Srv.Config.AirTempID = aid
	c.Srv.Config.SoilTempID = sid

	c.Srv.Config.WriteToFile("config.json")
}

// LogInfo is used to log information messages for this controller.
func (c *ConfigController) LogInfo(v ...interface{}) {
	a := fmt.Sprint(v...)
	logger.Info("ConfigController: ", a)
}

// LogError is used to log error messages for this controller.
func (c *ConfigController) LogError(v ...interface{}) {
	a := fmt.Sprint(v...)
	logger.Error("ConfigController: ", a)
}
