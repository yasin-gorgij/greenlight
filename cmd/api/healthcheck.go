package main

import (
	"net/http"
)

func (app *application) healthcheckHandler(resp http.ResponseWriter, req *http.Request) {
	env := envelope{
		"healthcheck": map[string]any{
			"status": "available",
			"system_info": map[string]string{
				"environment": app.config.env,
				"version":     version,
			},
		},
	}

	err := app.writeJSON(resp, http.StatusOK, env, nil)
	if err != nil {
		app.serverErrorResponse(resp, req, err)
	}
}
