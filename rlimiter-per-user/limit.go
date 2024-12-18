package main

import (
	"encoding/json"
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type Client struct {
	limiter     *rate.Limiter
	lastVisited time.Time
}

var clients = make(map[string]*Client)
var mu sync.Mutex

func CreateRateLimiterFromConfig(config *RateLimiterConfig) *rate.Limiter {
	limit := rate.Limit(config.Rate)
	return rate.NewLimiter(limit, config.Burst)
}

func InitializeRateLimiter(configPath string) (*rate.Limiter, error) {
	config, err := LoadRateLimiterConfig(configPath)
	if err != nil {
		return nil, err
	}
	return CreateRateLimiterFromConfig(config), nil
}

func perUserRateLimiter(next func(writer http.ResponseWriter, request *http.Request), config *RateLimiterConfig) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientIp, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			return
		}
		/* We need to make this thread safe caus clients' is a global
		map & can be accessed by other calls.. */
		mu.Lock()
		if _, exists := clients[clientIp]; !exists {
			clients[clientIp] = &Client{limiter: CreateRateLimiterFromConfig(config)}
		}
		clients[clientIp].lastVisited = time.Now()
		mu.Unlock()
		if !clients[clientIp].limiter.Allow() {
			message := Message{Status: "Request Failed",
				Body: "Too many requests",
			}
			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(&message)
		} else {
			next(w, r)
		}
	})
}
