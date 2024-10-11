package main

import (
	"fmt"
	"net/http"
)

func (app *application) createMovieHandler(resp http.ResponseWriter, req *http.Request) {
	fmt.Fprintln(resp, "create a new movie")
}

func (app *application) showMovieHandler(resp http.ResponseWriter, req *http.Request) {
	id, err := app.readIDParam(req)
	if err != nil || id < 1 {
		http.NotFound(resp, req)
		return
	}

	fmt.Fprintf(resp, "show the details of movie %d\n", id)
}
