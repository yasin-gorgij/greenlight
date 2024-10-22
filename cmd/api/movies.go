package main

import (
	"errors"
	"fmt"
	"greenlight/internal/data"
	"greenlight/internal/validator"
	"net/http"
)

func (app *application) createMovieHandler(resp http.ResponseWriter, req *http.Request) {
	var input struct {
		Title   string   `json:"title"`
		Year    int32    `json:"year"`
		Runtime string   `json:"runtime"`
		Genres  []string `json:"genres"`
	}

	err := app.readJSON(resp, req, &input)
	if err != nil {
		app.badRequestResponse(resp, req, err)
		return
	}

	movie := &data.Movie{
		Title:   input.Title,
		Year:    input.Year,
		Runtime: input.Runtime,
		Genres:  input.Genres,
	}

	v := validator.New()
	if data.ValidateMovie(v, movie); !v.Valid() {
		app.failedValidationResponse(resp, req, v.Errors)
		return
	}

	err = app.models.Movies.Insert(movie)
	if err != nil {
		app.serverErrorResponse(resp, req, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/movies/%d", movie.ID))

	err = app.writeJSON(resp, http.StatusCreated, envelope{"movie": movie}, headers)
	if err != nil {
		app.serverErrorResponse(resp, req, err)
	}
}

func (app *application) showMovieHandler(resp http.ResponseWriter, req *http.Request) {
	id, err := app.readIDParam(req)
	if err != nil || id < 1 {
		app.notFoundErrorRespone(resp, req)
		return
	}

	movie, err := app.models.Movies.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundErrorRespone(resp, req)
		default:
			app.serverErrorResponse(resp, req, err)
		}

		return
	}

	err = app.writeJSON(resp, http.StatusOK, envelope{"movie": movie}, nil)
	if err != nil {
		app.serverErrorResponse(resp, req, err)
	}
}

func (app *application) updateMovieHandler(resp http.ResponseWriter, req *http.Request) {
	id, err := app.readIDParam(req)
	if err != nil {
		app.notFoundErrorRespone(resp, req)
		return
	}

	movie, err := app.models.Movies.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundErrorRespone(resp, req)
		default:
			app.serverErrorResponse(resp, req, err)
		}

		return
	}

	var input struct {
		Title   string   `json:"title"`
		Year    int32    `json:"year"`
		Runtime string   `json:"runtime"`
		Genres  []string `json:"genres"`
	}

	err = app.readJSON(resp, req, &input)
	if err != nil {
		app.badRequestResponse(resp, req, err)
		return
	}

	movie.Title = input.Title
	movie.Year = input.Year
	movie.Runtime = input.Runtime
	movie.Genres = input.Genres

	v := validator.New()
	if data.ValidateMovie(v, movie); !v.Valid() {
		app.failedValidationResponse(resp, req, v.Errors)
		return
	}

	err = app.models.Movies.Update(movie)
	if err != nil {
		app.serverErrorResponse(resp, req, err)
		return
	}

	err = app.writeJSON(resp, http.StatusOK, envelope{"movie": movie}, nil)
	if err != nil {
		app.serverErrorResponse(resp, req, err)
	}
}

func (app *application) deleteMovieHandler(resp http.ResponseWriter, req *http.Request) {
	id, err := app.readIDParam(req)
	if err != nil {
		app.notFoundErrorRespone(resp, req)
		return
	}

	err = app.models.Movies.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundErrorRespone(resp, req)
		default:
			app.serverErrorResponse(resp, req, err)
		}

		return
	}

	err = app.writeJSON(resp, http.StatusOK, envelope{"message": "The movie successfully deleted."}, nil)
	if err != nil {
		app.serverErrorResponse(resp, req, err)
	}
}
