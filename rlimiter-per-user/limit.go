package main

import (
	"encoding/json"
	"net/http"
	"time"

	"golang.org/x/time/rate"
)

type Client struct {
	limter      *rate.Limiter
	lastVisited time.Time
}

// We need to have a map of clients

// globalMapOfClients := make(map[string]Client)

globalMapOfClients := make(map[string]Client)


func perUserRateLimiter(next func(w http.ResponseWriter, r *http.Request)) http.Handler {
	limiter := rate.NewLimiter(2, 4)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !limiter.Allow() {
			message := Message{Status: "Request Failed",
				Body: "Too many requests",
			}
			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(&message)
			return
		} else {
			next(w, r)
		}
	})
}
