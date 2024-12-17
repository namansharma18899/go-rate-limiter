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

// We need to have a map of clients
// globalMapOfClients := make(map[string]Client)

var clients = make(map[string]*Client)

func perUserRateLimiter(next func(writer http.ResponseWriter, request *http.Request)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientIp, _, err := net.SplitHostPort(r.RemoteAddr)
		var mu sync.Mutex
		if err != nil {
			return
		}
		/* We need to make this thread safe caus clients' is a global
		map & can be accessed by other calls.. */
		mu.Lock()
		if _, exists := clients[clientIp]; !exists {
			clients[clientIp] = &Client{limiter: rate.NewLimiter(2, 40)}
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
