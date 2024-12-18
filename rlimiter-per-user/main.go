package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

var configPath = "./config.json"

type Message struct {
	Status string
	Body   string
}

func endpointHnadler(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	message := Message{"Successfull", "jkjk"}
	fmt.Print(message)
	err := json.NewEncoder(writer).Encode(&message)
	if err != nil {
		return
	}
}

func main() {
	config, err := LoadRateLimiterConfig(configPath)
	if err != nil {
		fmt.Errorf("Could not load Rate limiter config !!")
	}
	http.Handle("/ping", perUserRateLimiter(endpointHnadler, config))
	err = http.ListenAndServe(":8080", nil)
	fmt.Print("Startee Listening...")
	if err == nil {
		fmt.Errorf("Error listening on port 8080")
	}

}
