package main

import (
	"context"
	"greenlight/internal/data"
	"net/http"
)

type contextKey string

const userContextKey = contextKey("user")

func (app *application) contextSetUser(req *http.Request, user *data.User) *http.Request {
	ctx := context.WithValue(req.Context(), userContextKey, user)
	return req.WithContext(ctx)
}

func (app *application) contextGetUser(r *http.Request) *data.User {
	user, ok := r.Context().Value(userContextKey).(*data.User)
	if !ok {
		panic("missing user value in request context")
	}

	return user
}
