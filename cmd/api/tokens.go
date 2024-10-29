package main

import (
	"errors"
	"greenlight/internal/data"
	"greenlight/internal/validator"
	"net/http"
	"time"
)

func (app *application) createActivationTokenHandler(resp http.ResponseWriter, req *http.Request) {
	var input struct {
		Email string `json:"email"`
	}

	err := app.readJSON(resp, req, &input)
	if err != nil {
		app.badRequestResponse(resp, req, err)
		return
	}

	v := validator.New()
	if data.ValidateEmail(v, input.Email); !v.Valid() {
		app.failedValidationResponse(resp, req, v.Errors)
		return
	}

	user, err := app.models.Users.GetByEmail(input.Email)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			v.AddError("email", "no matching email address found")
			app.failedValidationResponse(resp, req, v.Errors)
		default:
			app.serverErrorResponse(resp, req, err)
		}

		return
	}
	if user.Activated {
		v.AddError("email", "user has already been activated")
		app.failedValidationResponse(resp, req, v.Errors)
		return
	}

	token, err := app.models.Tokens.New(user.ID, 3*24*time.Hour, data.ScopeActivation)
	if err != nil {
		app.serverErrorResponse(resp, req, err)
		return
	}

	app.background(func() {
		data := map[string]any{
			"activationToken": token.Plaintext,
		}

		err = app.mailer.Send(user.Email, "token_activation.tmpl.html", data)
		if err != nil {
			app.logger.Error(err.Error())
		}
	})

	env := envelope{"message": "an email will be sent to you containing activation instructions"}

	err = app.writeJSON(resp, http.StatusAccepted, env, nil)
	if err != nil {
		app.serverErrorResponse(resp, req, err)
	}
}