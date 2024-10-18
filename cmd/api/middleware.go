package main

import (
	"fmt"
	"net/http"
)

func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				resp.Header().Set("Connection", "close")
				app.serverErrorResponse(resp, req, fmt.Errorf("%s", err))
			}
		}()

		next.ServeHTTP(resp, req)
	})
}
