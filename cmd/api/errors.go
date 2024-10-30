package main

import (
	"fmt"
	"log/slog"
	"net/http"
)

func (app *application) logError(req *http.Request, err error) {
	app.logger.Error(err.Error(), slog.String("method", req.Method), slog.String("uri", req.URL.RequestURI()))
}

func (app *application) errorResponse(resp http.ResponseWriter, req *http.Request, status int, message any) {
	err := app.writeJSON(resp, status, envelope{"error": message}, nil)
	if err != nil {
		app.logError(req, err)
		resp.WriteHeader(http.StatusInternalServerError)
	}
}

func (app *application) serverErrorResponse(resp http.ResponseWriter, req *http.Request, err error) {
	app.logError(req, err)
	app.errorResponse(resp, req, http.StatusInternalServerError, "The server encountered a problem and could not process your request")
}

func (app *application) notFoundErrorRespone(resp http.ResponseWriter, req *http.Request) {
	app.errorResponse(resp, req, http.StatusNotFound, "The requested resource could not be found")
}

func (app *application) methodNotAllowedErrorResponse(resp http.ResponseWriter, req *http.Request) {
	message := fmt.Sprintf("The %s method is not supported for this resource", req.Method)
	app.errorResponse(resp, req, http.StatusMethodNotAllowed, message)
}

func (app *application) badRequestResponse(resp http.ResponseWriter, req *http.Request, err error) {
	app.errorResponse(resp, req, http.StatusBadRequest, err.Error())
}

func (app *application) failedValidationResponse(resp http.ResponseWriter, req *http.Request, errors map[string]string) {
	app.errorResponse(resp, req, http.StatusUnprocessableEntity, errors)
}

func (app *application) editConflictResponse(resp http.ResponseWriter, req *http.Request) {
	message := "unable to update the record due to an edit conflict, please try again"
	app.errorResponse(resp, req, http.StatusConflict, message)
}

func (app *application) rateLimitExceededResponse(w http.ResponseWriter, r *http.Request) {
	message := "rate limit exceeded"
	app.errorResponse(w, r, http.StatusTooManyRequests, message)
}

func (app *application) invalidCredentialsResponse(resp http.ResponseWriter, req *http.Request) {
	message := "invalid authentication credentials"
	app.errorResponse(resp, req, http.StatusUnauthorized, message)
}

func (app *application) invalidAuthenticationTokenResponse(resp http.ResponseWriter, req *http.Request) {
	resp.Header().Set("WWW-Authenticate", "Bearer")
	message := "invalid or missing authentication token"
	app.errorResponse(resp, req, http.StatusUnauthorized, message)
}

func (app *application) authenticationRequiredResponse(resp http.ResponseWriter, req *http.Request) {
	message := "you must be authenticated to access this resource"
	app.errorResponse(resp, req, http.StatusUnauthorized, message)
}

func (app *application) inactiveAccountResponse(resp http.ResponseWriter, req *http.Request) {
	message := "your user account must be activated to access this resource"
	app.errorResponse(resp, req, http.StatusForbidden, message)
}

func (app *application) notPermittedResponse(resp http.ResponseWriter, req *http.Request) {
	message := "your user account doesn't have the necessary permissions to access this resource"
	app.errorResponse(resp, req, http.StatusForbidden, message)
}
