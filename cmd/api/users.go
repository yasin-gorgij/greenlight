package main

import (
	"errors"
	"greenlight/internal/data"
	"greenlight/internal/validator"
	"net/http"
)

func (app *application) registerUserHandler(resp http.ResponseWriter, req *http.Request) {
	var input struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := app.readJSON(resp, req, &input)
	if err != nil {
		app.badRequestResponse(resp, req, err)
		return
	}

	user := &data.User{
		Name:      input.Name,
		Email:     input.Email,
		Activated: false,
	}

	err = user.Password.Set(input.Password)
	if err != nil {
		app.serverErrorResponse(resp, req, err)
		return
	}

	v := validator.New()
	if data.ValidateUser(v, user); !v.Valid() {
		app.failedValidationResponse(resp, req, v.Errors)
		return
	}

	err = app.models.Users.Insert(user)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDuplicateEmail):
			v.AddError("email", "a user with this email address already exists")
			app.failedValidationResponse(resp, req, v.Errors)
		default:
			app.serverErrorResponse(resp, req, err)
		}
		return
	}

	err = app.writeJSON(resp, http.StatusCreated, envelope{"user": user}, nil)
	if err != nil {
		app.serverErrorResponse(resp, req, err)
	}
}
