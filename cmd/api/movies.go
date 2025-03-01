package main

import (
	"fmt"
	"net/http"
	"errors"
	"greenlight.nursultandias.net/internal/data"
	"greenlight.nursultandias.net/internal/validator"
)

func (app *application) createMovieHandler(response http.ResponseWriter, request *http.Request) {
	
	// Declare an anonymous struct to hold the information that we expect to be in the
	// HTTP request body (note that the field names and types in the struct are a subset
	// of the Movie struct that we created earlier). This struct will be our *target decode destination*.
	var input struct {
		Title	string			`json:"title"`
		Year	int32			`json:"year"`
		Runtime	data.Runtime	`json:"runtime"`
		Genres	[]string		`json:"genres"`
	}

	// Use the new readJSON() helper to decode the request body into the input struct.
	// If this returns an error we send the client the error message along with a 400
	// Bad Request status code, just like before.
	err := app.readJSON(response, request, &input)
	if err != nil {
		app.badRequestResponse(response, request, err)
		return 
	}

	// Copy the values from the input struct to a new Movie struct.
	// Note that the movie variable contains a *pointer* to a Movie struct.
	movie := &data.Movie{ 
		Title: input.Title,
		Year: input.Year,
		Runtime: input.Runtime,
		Genres: input.Genres,
	}

	// Initialize a new Validator instance.
	v := validator.New()

	// Call the ValidateMovie() function and return a response containing the errors if // any of the checks fail.
	if data.ValidateMovie(v, movie); !v.Valid() {
		app.failedValidationResponse(response, request, v.Errors)
		return
	}

	// Call the Insert() method on our movies model, passing in a pointer to the
	// validated movie struct. This will create a record in the database and update the
	// movie struct with the system-generated information.
	err = app.models.Movies.Insert(movie)
	if err != nil {
		app.serverErrorResponse(response, request, err)
		return
	}

	// When sending a HTTP response, we want to include a Location header to let the
	// client know which URL they can find the newly-created resource at. We make an
	// empty http.Header map and then use the Set() method to add a new Location header,
	// interpolating the system-generated ID for our new movie in the URL.
	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/movies/%d", movie.ID))

	// Write a JSON response with a 201 Created status code, the movie data in the
	// response body, and the Location header.
	err = app.writeJSON(response, http.StatusCreated, envelope{"movie": movie}, headers)
	if err != nil {
		app.serverErrorResponse(response, request, err)
	}
}

func (app *application) showMovieHandler(response http.ResponseWriter, request *http.Request) {
	id, err := app.readIDParam(request)
	if err != nil {
		app.notFoundResponse(response, request)
		return
	}

	// Call the Get() method to fetch the data for a specific movie. We also need to
	// use the errors.Is() function to check if it returns a data.ErrRecordNotFound
	// error, in which case we send a 404 Not Found response to the client.
	movie, err := app.models.Movies.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(response, request)
		default:
			app.serverErrorResponse(response, request, err)
		}
		return
	}
	

	err = app.writeJSON(response, http.StatusOK, envelope{"movie": movie}, nil)
	if err != nil {
		app.serverErrorResponse(response, request, err)
	}
}

func (app *application) updateMovieHandler(response http.ResponseWriter, request *http.Request) {
	// Extract the movie ID from the URL.
	id, err := app.readIDParam(request)
	if err != nil {
		app.notFoundResponse(response, request)
		return
	}

	// Fetch the existing movie record from the database, sending a 404 Not Found
	// response to the client if we couldn't find a matching record.
	movie, err := app.models.Movies.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(response, request) 
		default:
			app.serverErrorResponse(response, request, err) }
		return
	}

	// Declare an input struct to hold the expected data from the client.
	// To support partial updates, use pointers for the Title, Year and Runtime fields.
	var input struct {
		Title		*string			`json:"title"`		// This will be nil if there is no corresponding key in the JSON.
		Year		*int32			`json:"year"`		// Likewise...
		Runtime		*data.Runtime	`json:"runtime"`	// Likewise...
		Genres		[]string		`json:"genres"`		// We don't need to change this because slices already have the zero-value nil.
	}

	// Read the JSON request body data into the input struct.
	err = app.readJSON(response, request, &input)
	if err != nil {
		app.badRequestResponse(response, request, err)
		return
	}

	// If the input.Title value is nil then we know that no corresponding "title" key/
	// value pair was provided in the JSON request body. So we move on and leave the
	// movie record unchanged. Otherwise, we update the movie record with the new title
	// value. Importantly, because input.Title is a now a pointer to a string, we need
	// to dereference the pointer using the * operator to get the underlying value
	// before assigning it to our movie record.
	if input.Title != nil {
		movie.Title = *input.Title
	}

	// We also do the same for the other fields in the input struct.
	if input.Year != nil {
		movie.Year = *input.Year
	}
	if input.Runtime != nil {
		movie.Runtime = *input.Runtime
	}
	if input.Genres != nil {
		movie.Genres = input.Genres // Note that we don't need to dereference a slice.
	}

	// Validate the updated movie record, sending the client a 422 Unprocessable Entity
	// response if any checks fail.
	v := validator.New()

	if data.ValidateMovie(v, movie); !v.Valid() {
		app.failedValidationResponse(response, request, v.Errors)
		return
	}

	// Pass the updated movie record to our new Update() method.
	err = app.models.Movies.Update(movie)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(response, request)
		default:
			app.serverErrorResponse(response, request, err)
		}
		return
	}

	// Write the updated movie record in a JSON response.
	err = app.writeJSON(response, http.StatusOK, envelope{"movie": movie}, nil)
	if err != nil {
		app.serverErrorResponse(response, request, err)
	}
}

func (app *application) deleteMovieHandler(response http.ResponseWriter, request *http.Request) {
	// Extract the movie ID from the URL.
	id, err := app.readIDParam(request)
	if err != nil {
		app.notFoundResponse(response, request)
		return
	}

	// Delete the movie from the database, sending a 404 Not Found response to the
	// client if there isn't a matching record.
	err = app.models.Movies.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(response, request)
		default:
			app.serverErrorResponse(response, request, err)
		}
		return
	}

	// Return a 200 OK status code along with a success message.
	err = app.writeJSON(response, http.StatusOK, envelope{"message": "movie successfully deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(response, request, err)
	}
}

func (app *application) listMoviesHandler(response http.ResponseWriter, request *http.Request) {
	// To keep things consistent with our other handlers, we'll define an input struct
	// to hold the expected values from the request query string.
	// Embed the new Filters struct.
	var input struct {
		Title		string
		Genres		[]string
		data.Filters
	}

	// Initialize a new Validator instance.
	v := validator.New()

	// Call r.URL.Query() to get the url.Values map containing the query string data.
	qs := request.URL.Query()

	// Use our helpers to extract the title and genres query string values, falling back
	// to defaults of an empty string and an empty slice respectively if they are not
	// provided by the client.
	input.Title = app.readString(qs, "title", "")
	input.Genres = app.readCSV(qs, "genres", []string{})

	// Get the page and page_size query string values as integers. Notice that we set
	// the default page value to 1 and default page_size to 20, and that we pass the
	// validator instance as the final argument here.
	input.Filters.Page = app.readInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readInt(qs, "page_size", 20, v)

	// Extract the sort query string value, falling back to "id" if it is not provided // by the client (which will imply a ascending sort on movie ID).
	input.Filters.Sort = app.readString(qs, "sort", "id")
	// Add the supported sort values for this endpoint to the sort safelist.
	input.Filters.SortSafelist = []string{"id", "title", "year", "runtime", "-id", "-title", "-year", "-runtime"}

	// Execute the validation checks on the Filters struct and send a response
	// containing the errors if necessary.
	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(response, request, v.Errors)
		return
	}

	// Call the GetAll() method to retrieve the movies, passing in the various filter parameters.
	movies, metadata ,err := app.models.Movies.GetAll(input.Title, input.Genres, input.Filters)
	if err != nil {
		app.serverErrorResponse(response, request, err)
		return
	}

	// Send a JSON response containing the movie data.
	err = app.writeJSON(response, http.StatusOK, envelope{"movies": movies, "metadata" : metadata}, nil)
	if err != nil {
		app.serverErrorResponse(response, request, err)
	}
}