package main

import (
	"errors"
	"greenlight/internal/data"
	"greenlight/internal/validator"
	"net/http"
	"time"
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

	err = app.models.Permissions.AddForUser(user.ID, "movies:read")
	if err != nil {
		app.serverErrorResponse(resp, req, err)
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
			"userID":          user.ID,
		}

		err = app.mailer.Send(user.Email, "user_welcome.tmpl.html", data)
		if err != nil {
			app.logger.Error(err.Error())
		}
	})

	err = app.writeJSON(resp, http.StatusAccepted, envelope{"user": user}, nil)
	if err != nil {
		app.serverErrorResponse(resp, req, err)
	}
}

func (app *application) activateUserHandler(resp http.ResponseWriter, req *http.Request) {
	var input struct {
		TokenPlaintext string `json:"token"`
	}

	err := app.readJSON(resp, req, &input)
	if err != nil {
		app.badRequestResponse(resp, req, err)
		return
	}

	v := validator.New()
	if data.ValidateTokenPlaintext(v, input.TokenPlaintext); !v.Valid() {
		app.failedValidationResponse(resp, req, v.Errors)
		return
	}

	user, err := app.models.Users.GetForToken(data.ScopeActivation, input.TokenPlaintext)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			v.AddError("token", "invalid or expired activation token")
			app.failedValidationResponse(resp, req, v.Errors)
		default:
			app.serverErrorResponse(resp, req, err)
		}

		return
	}

	user.Activated = true

	err = app.models.Users.Update(user)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(resp, req)
		default:
			app.serverErrorResponse(resp, req, err)
		}

		return
	}

	err = app.models.Tokens.DeleteAllForUser(data.ScopeActivation, user.ID)
	if err != nil {
		app.serverErrorResponse(resp, req, err)
		return
	}

	err = app.writeJSON(resp, http.StatusOK, envelope{"user": user}, nil)
	if err != nil {
		app.serverErrorResponse(resp, req, err)
	}
}

func (app *application) updateUserPasswordHandler(resp http.ResponseWriter, req *http.Request) {
	var input struct {
		Password       string `json:"password"`
		TokenPlaintext string `json:"token"`
	}

	err := app.readJSON(resp, req, &input)
	if err != nil {
		app.badRequestResponse(resp, req, err)
		return
	}

	v := validator.New()

	data.ValidatePasswordPlaintext(v, input.Password)
	data.ValidateTokenPlaintext(v, input.TokenPlaintext)

	if !v.Valid() {
		app.failedValidationResponse(resp, req, v.Errors)
		return
	}

	user, err := app.models.Users.GetForToken(data.ScopePasswordReset, input.TokenPlaintext)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			v.AddError("token", "invalid or expired password reset token")
			app.failedValidationResponse(resp, req, v.Errors)
		default:
			app.serverErrorResponse(resp, req, err)
		}
		return
	}

	err = user.Password.Set(input.Password)
	if err != nil {
		app.serverErrorResponse(resp, req, err)
		return
	}

	err = app.models.Users.Update(user)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(resp, req)
		default:
			app.serverErrorResponse(resp, req, err)
		}
		return
	}

	err = app.models.Tokens.DeleteAllForUser(data.ScopePasswordReset, user.ID)
	if err != nil {
		app.serverErrorResponse(resp, req, err)
		return
	}

	env := envelope{"message": "your password was successfully reset"}

	err = app.writeJSON(resp, http.StatusOK, env, nil)
	if err != nil {
		app.serverErrorResponse(resp, req, err)
	}
}
