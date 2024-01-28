package router

import (
	"net/http"
	"sync"
	"time"
)

type (
	RateLimiter struct {
		config   RateLimiterConfig
		requests map[string]*request
		mutex    sync.Mutex
		logger   Logger
	}

	RateLimiterConfig struct {
		Duration  time.Duration
		Limit     uint8
		Extractor Extractor
	}

	Extractor func(r *http.Request) string

	Logger func(identifier, url string)

	request struct {
		time  time.Time
		count uint8
	}
)

func NewRateLimiterMiddleware(config RateLimiterConfig, logger Logger) *RateLimiter {
	return &RateLimiter{
		config:   config,
		requests: make(map[string]*request),
		logger:   logger,
	}
}

func (rateLimiter *RateLimiter) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rateLimiter.mutex.Lock()

		defer rateLimiter.mutex.Unlock()

		var identifier = rateLimiter.config.Extractor(r)

		if _, exists := rateLimiter.requests[identifier]; !exists {
			rateLimiter.requests[identifier] = &request{
				time:  time.Now(),
				count: 1,
			}

			next.ServeHTTP(w, r)

			return
		}

		req := rateLimiter.requests[identifier]

		if time.Since(req.time) > rateLimiter.config.Duration {
			req.time = time.Now()
			req.count = 1

			next.ServeHTTP(w, r)

			return
		}

		if req.count < rateLimiter.config.Limit {
			req.count++

			next.ServeHTTP(w, r)

			return
		}

		rateLimiter.logger(identifier, r.URL.String())

		w.WriteHeader(http.StatusTooManyRequests)

		return
	})
}
