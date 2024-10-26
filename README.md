```go
package main

import (
	"github.com/amirrezam75/go-router"
	m "github.com/amirrezam75/go-router/middlewares"
	"net/http"
)

func setupRoutes() *router.Router {
	var r = router.NewRouter()

	r.Middleware(middlewares.CorsPolicy{})

	var rateLimiterLogger = func(identifier, url string) {
		services.LogService{}.Info(map[string]string{
			"message":    "Too many requests.",
			"identifier": identifier,
			"url":        url,
		})
	}

	var rateLimiterConfig = m.RateLimiterConfig{
		Duration: time.Minute,
		Limit:    15,
		Extractor: func(r *http.Request) string {
			var user = services.ContextService{}.GetUser(r.Context())

			if user == nil {
				return r.RemoteAddr
			}

			return user.Id
		},
	}

	var rateLimiterMiddleware = m.NewRateLimiterMiddleware(rateLimiterConfig, rateLimiterLogger)

	r.Post("/lobbies", lobbyHandler.Create).
		Middleware(authMiddleware).
		Middleware(rateLimiterMiddleware)

	return r
}

```

# Middlewares

```go
package middlewares

import (
	"net/http"
	"os"
)

type CorsPolicy struct {
}

func (cp CorsPolicy) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", os.Getenv("FRONTEND_URL"))
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, OPTIONS")
		w.Header().Set("Access-Control-Max-Age", "3600")

		// Handles preflight requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

```