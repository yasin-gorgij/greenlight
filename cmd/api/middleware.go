package main

import (
	"errors"
	"expvar"
	"fmt"
	"greenlight/internal/data"
	"greenlight/internal/validator"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type metricsResponseWriter struct {
	wrapped       http.ResponseWriter
	statusCode    int
	headerWritten bool
}

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
	type client struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	}

	var (
		mu      sync.Mutex
		clients = make(map[string]*client)
	)

	go func() {
		for {
			time.Sleep(time.Minute)
			mu.Lock()

			for ip, client := range clients {
				if time.Since(client.lastSeen) > 3*time.Minute {
					delete(clients, ip)
				}
			}

			mu.Unlock()
		}
	}()

	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		if app.config.limiter.enabled {

			ip, _, err := net.SplitHostPort(req.RemoteAddr)
			if err != nil {
				app.serverErrorResponse(resp, req, err)
				return
			}

			mu.Lock()

			if _, found := clients[ip]; !found {
				clients[ip] = &client{limiter: rate.NewLimiter(2, 4)}
			}
			clients[ip].lastSeen = time.Now()

			if !clients[ip].limiter.Allow() {
				mu.Unlock()
				app.rateLimitExceededResponse(resp, req)
				return
			}

			mu.Unlock()
		}

		next.ServeHTTP(resp, req)
	})
}

func (app *application) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		resp.Header().Add("Vary", "Authorization")

		authorizationHeader := req.Header.Get("Authorization")
		if authorizationHeader == "" {
			req = app.contextSetUser(req, data.AnonymousUser)
			next.ServeHTTP(resp, req)
			return
		}

		headerParts := strings.Split(authorizationHeader, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			app.invalidAuthenticationTokenResponse(resp, req)
			return
		}

		token := headerParts[1]
		v := validator.New()
		if data.ValidateTokenPlaintext(v, token); !v.Valid() {
			app.invalidAuthenticationTokenResponse(resp, req)
			return
		}

		user, err := app.models.Users.GetForToken(data.ScopeAuthentication, token)
		if err != nil {
			switch {
			case errors.Is(err, data.ErrRecordNotFound):
				app.invalidAuthenticationTokenResponse(resp, req)
			default:
				app.serverErrorResponse(resp, req, err)

			}
			return
		}

		req = app.contextSetUser(req, user)

		next.ServeHTTP(resp, req)
	})
}

func (app *application) requirePermission(code string, next http.HandlerFunc) http.HandlerFunc {
	fn := func(resp http.ResponseWriter, req *http.Request) {
		user := app.contextGetUser(req)

		permissions, err := app.models.Permissions.GetAllForUser(user.ID)
		if err != nil {
			app.serverErrorResponse(resp, req, err)
			return
		}

		if !permissions.Include(code) {
			app.notPermittedResponse(resp, req)
			return
		}

		next.ServeHTTP(resp, req)
	}

	return app.requireActivatedUser(fn)
}

func (app *application) requireActivatedUser(next http.HandlerFunc) http.HandlerFunc {
	fn := http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		user := app.contextGetUser(req)
		if !user.Activated {
			app.inactiveAccountResponse(resp, req)
			return
		}

		next.ServeHTTP(resp, req)
	})

	return app.requireAuthenticatedUser(fn)
}

func (app *application) requireAuthenticatedUser(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		user := app.contextGetUser(req)
		if user.IsAnonymous() {
			app.authenticationRequiredResponse(resp, req)
			return
		}

		next.ServeHTTP(resp, req)
	})
}

func (app *application) enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		resp.Header().Add("Vary", "Origin")
		resp.Header().Add("Vary", "Access-Control-Request-Method")

		origin := req.Header.Get("Origin")
		if origin != "" {
			for i := range app.config.cors.trustedOrigins {
				if origin == app.config.cors.trustedOrigins[i] {
					resp.Header().Set("Access-Control-Allow-Origin", origin)
					resp.Header().Set("Access-Control-Allow-Credentials", "true")

					if req.Method == http.MethodOptions && req.Header.Get("Access-Control-Request-Method") != "" {
						resp.Header().Set("Access-Control-Allow-Methods", "OPTIONS, PUT, PATCH, DELETE")
						resp.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
						resp.WriteHeader(http.StatusOK)

						return

					}
					break
				}
			}
		}

		next.ServeHTTP(resp, req)
	})
}

func newMetricsResponseWriter(resp http.ResponseWriter) *metricsResponseWriter {
	return &metricsResponseWriter{
		wrapped:    resp,
		statusCode: http.StatusOK,
	}
}

func (metricsResp *metricsResponseWriter) Header() http.Header {
	return metricsResp.wrapped.Header()
}

func (metricsResp *metricsResponseWriter) WriteHeader(statusCode int) {
	metricsResp.wrapped.WriteHeader(statusCode)
	if !metricsResp.headerWritten {
		metricsResp.statusCode = statusCode
		metricsResp.headerWritten = true
	}
}

func (metricsResp *metricsResponseWriter) Write(b []byte) (int, error) {
	metricsResp.headerWritten = true
	return metricsResp.wrapped.Write(b)
}

func (metricsResp *metricsResponseWriter) Unwrap() http.ResponseWriter {
	return metricsResp.wrapped
}

func (app *application) metrics(next http.Handler) http.Handler {
	var (
		totalRequestReceived            = expvar.NewInt("total_request_received")
		totalResponsesSent              = expvar.NewInt("total_responses_sent")
		totalProcessingTimeMicroseconds = expvar.NewInt("total_processing_time_μs")
	)

	totalResponsesSentByStatus := expvar.NewMap("total_responses_sent_by_status")

	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		start := time.Now()
		totalRequestReceived.Add(1)

		metricsResp := newMetricsResponseWriter(resp)
		next.ServeHTTP(metricsResp, req)

		totalResponsesSent.Add(1)
		totalResponsesSentByStatus.Add(strconv.Itoa(metricsResp.statusCode), 1)

		duration := time.Since(start).Microseconds()
		totalProcessingTimeMicroseconds.Add(duration)
	})
}
