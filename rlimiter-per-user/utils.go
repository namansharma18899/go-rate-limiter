package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type RateLimiterConfig struct {
	Rate  float64 `json:"rate"`  // Tokens per second
	Burst int     `json:"burst"` // Maximum burst of tokens
}

func LoadRateLimiterConfig(filepath string) (*RateLimiterConfig, error) {
	configFile, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %v", err)
	}
	var config RateLimiterConfig
	err = json.Unmarshal(configFile, &config)
	if err != nil {
		return nil, fmt.Errorf("error parsing config file: %v", err)
	}
	return &config, nil
}
