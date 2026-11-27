package main

import (
	"os"
	"strings"
)

const (
	defaultBaseURL = "http://127.0.0.1:9090/v1"
	defaultModel   = "gemma-4-e4b-it-4bit"
)

type config struct {
	BaseURL string
	Model   string
	APIKey  string
}

func loadConfig() config {
	return config{
		BaseURL: strings.TrimRight(envOr("KOMIT_BASE_URL", defaultBaseURL), "/"),
		Model:   envOr("KOMIT_MODEL", defaultModel),
		APIKey:  os.Getenv("KOMIT_API_KEY"),
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
