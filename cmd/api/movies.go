package main

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

func (app *application) createMovieHandler(resp http.ResponseWriter, req *http.Request) {
	fmt.Fprintln(resp, "create a new movie")
}

func (app *application) showMovieHandler(resp http.ResponseWriter, req *http.Request) {
	params := httprouter.ParamsFromContext(req.Context())
	id, err := strconv.ParseInt(params.ByName("id"), 10, 64)
	if err != nil || id < 1 {
		http.NotFound(resp, req)
		return
	}

	fmt.Fprintf(resp, "show the details of movie %d\n", id)
}
