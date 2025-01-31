package main

import (
	"net/http"
)

func (app *application) healthcheckHandler(response http.ResponseWriter, request *http.Request) {
	// Create a map which holds the information that we want to send in the response. 
	env := envelope{
		"status": "available", 
		"system_info": map[string]string{
						"environment": app.config.env,
						"version": version,
					},
		}

	err := app.writeJSON(response, http.StatusOK, env, nil)
	if err != nil {
		app.serverErrorResponse(response, request, err)
	}
}