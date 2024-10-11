package main

import (
	"fmt"
	"net/http"
)

func (app *application) healthcheckHandler(resp http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(resp, "status: %s\n", "available")
	fmt.Fprintf(resp, "environment: %s\n", app.config.env)
	fmt.Fprintf(resp, "version: %s\n", version)
}
