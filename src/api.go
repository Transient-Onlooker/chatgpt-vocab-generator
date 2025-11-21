package main

import (
	"encoding/json"
	"os"
)

type APIConfig struct {
	APIKey string `json:"chatgpt_api_key"`
}

func loadAPIKey() (string, error) {
	data, err := os.ReadFile("api.json")
	if err != nil {
		return "", err
	}

	var config APIConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return "", err
	}

	return config.APIKey, nil
}
