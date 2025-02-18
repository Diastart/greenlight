package main

import (
	"fmt"
	"net/http"
)

// The logError() method is a genereric helper for logging an error message.
// Later we'll upgrade this to use structured logging, and record additional information
// about the request including the HTTP method and URL.
func (app *application) logError(request *http.Request, err error) {
	// Use the PrintError() method to log the error message, and include the current
	// request method and URL as properties in the log entry.
	app.logger.PrintError(err, map[string]string{
		"request_method":	request.Method,
		"request_url":		request.URL.String(),
	})
}

// The errorResponse() method is generic helper for sending JSON-formatted error
// messages to the client with a given status code. Note that we're using an interface{}
// type for the message parameter, rather than just a string type, as this gives us
// more flexibility over the values that we can include in the response.
func (app *application) errorResponse(response http.ResponseWriter, request *http.Request, status int, message interface{}) {
	env := envelope{"error": message}

	// Write the response using the writeJSON helper. If this happens to return an
	// error then log it, and fall back to sending the client an empty response
	// with a 500 Internal Server Error status code.
	err := app.writeJSON(response, status, env, nil)
	if err != nil {
		app.logError(request, err)
		response.WriteHeader(500)
	}
}

// The serverErrorResponse() method will be used when our application encounters an
// unexpected problem at runtime. It logs the detailed error message, then uses the
// errorResponse() helper to send a 500 Internal Server Error status code and JSON
// response (containing a generic error message) to the client.
func (app *application) serverErrorResponse(response http.ResponseWriter, request *http.Request, err error){
	app.logError(request, err)

	message := "the server ecnountered a problem and could not process your request"
	app.errorResponse(response, request, http.StatusInternalServerError, message)
}

// The notFoundResponse() method will be used to send a 404 Not Found status code and
// JSON response to the client.
func (app *application) notFoundResponse(response http.ResponseWriter, request *http.Request) {
	message := "the requested resource could not be found"
	app.errorResponse(response, request, http.StatusNotFound, message)
}

// The methodNotAllowedResponse() method will be used to send a 405 Method Not Allowed
// status code and JSON response to the client.
func (app *application) methodNotAllowedResponse(response http.ResponseWriter, request *http.Request) {
	message := fmt.Sprintf("the %s method is not supported for this resource", request.Method)
	app.errorResponse(response, request, http.StatusMethodNotAllowed, message)
}

func (app *application) badRequestResponse(response http.ResponseWriter, request *http.Request, err error) { 
	app.errorResponse(response, request, http.StatusBadRequest, err.Error())
}

// Note that the errors parameter here has the type map[string]string, which is exactly
// the same as the errors map contained in our Validator type.
func (app *application) failedValidationResponse(response http.ResponseWriter, request *http.Request, errors map[string]string) {
	app.errorResponse(response, request, http.StatusUnprocessableEntity, errors)
}

func (app *application) editConflictResponse(response http.ResponseWriter, request *http.Request) {
	message := "unable to update the record due to an edit conflict, please try again"
	app.errorResponse(response, request, http.StatusConflict, message)
}