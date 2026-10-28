package main

import (
	"log"
	"net/http"
)

func (app *application) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]string{
		"status":  "ok",
		"env":     app.config.env,
		"version": "0.0.0.1",
	}

	if err := writeJSON(w, http.StatusOK, data); err != nil {
		log.Print(err.Error())
	}
}
