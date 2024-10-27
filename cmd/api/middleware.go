package main

import (
	"fmt"
	"net/http"

	"golang.org/x/time/rate"
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

func (app *application) rateLimit(next http.Handler) http.Handler {
	limiter := rate.NewLimiter(2, 4)

	return http.HandleFunc(func(resp http.ResponseWriter, req *http.Request) {
		if !limiter.Allow() {
			app.rateLimitExceededResponse(resp, req)
			return
		}

		next.ServeHTTP(resp, req)
	})
}
