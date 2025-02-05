package main

import (
	"net/http"

	"github.com/VinicciusSantos/golang-microservices/logger-service/data"
)

type JSONPayload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

func (app *Config) WriteLog(w http.ResponseWriter, r *http.Request) {
	var requestPayload JSONPayload

	if err := app.readJSON(w, r, &requestPayload); err != nil {
		app.errorJSON(w, err)
		return
	}

	if err := app.Models.LogEntry.Insert(data.LogEntry{
		Name: requestPayload.Name,
		Data: requestPayload.Data,
	}); err != nil {
		app.errorJSON(w, err)
		return
	}

	app.writeJSON(w, http.StatusAccepted, jsonResponse{
		Error:   false,
		Message: "logged",
	})
}
