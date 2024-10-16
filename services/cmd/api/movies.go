package main

import (
	"fmt"
	"greenlight/internal/data"
	"net/http"
	"time"
)

func (app *application) createMovieHandler(resp http.ResponseWriter, req *http.Request) {
	fmt.Fprintln(resp, "create a new movie")
}

func (app *application) showMovieHandler(resp http.ResponseWriter, req *http.Request) {
	id, err := app.readIDParam(req)
	if err != nil || id < 1 {
		app.notFoundErrorRespone(resp, req)
		return
	}

	movie := data.Movie{
		ID:        id,
		CreatedAt: time.Now(),
		Title:     "Casablanca",
		Runtime:   "102 mins",
		Genres:    []string{"drama", "romance", "war"},
		Version:   1,
	}

	err = app.writeJSON(resp, http.StatusOK, envelope{"movie": movie}, nil)
	if err != nil {
		app.serverErrorResponse(resp, req, err)
	}
}
