package main

import (
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")
				app.serverErrorResponse(w, r, fmt.Errorf("%s", err))
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func (app *application) rateLimit(next http.Handler) http.Handler {
	type client struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	}

	limiters := struct {
		clients map[string]*client
		sync.Mutex
	}{
		clients: make(map[string]*client),
	}

	go func() {
		for {
			time.Sleep(1 * time.Minute)

			limiters.Lock()
			for ip, client := range limiters.clients {
				if time.Since(client.lastSeen) > 3*time.Minute {
					delete(limiters.clients, ip)
				}
			}
			limiters.Unlock()
		}
	}()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if app.config.limiter.enabled {
			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				app.serverErrorResponse(w, r, err)
				return
			}

			limiters.Lock()
			_, ok := limiters.clients[ip]
			if !ok {
				limiters.clients[ip] = &client{
					limiter: rate.NewLimiter(rate.Limit(app.config.limiter.rps), app.config.limiter.burst),
				}
			}
			limiters.clients[ip].lastSeen = time.Now()

			if !limiters.clients[ip].limiter.Allow() {
				app.rateLimitExceededResponse(w, r)
				limiters.Unlock()
				return
			}
			limiters.Unlock()
		}

		next.ServeHTTP(w, r)
	})
}
