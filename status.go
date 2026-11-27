package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

const modelsTimeout = 5 * time.Second

// runStatus prints the configuration and checks the server, model and
// generation. It exits 0 only when generation works.
func runStatus(args []string) int {
	cfg := loadConfig()

	fmt.Printf("Base URL   %s\n", orNotSet(cfg.BaseURL))
	fmt.Printf("Model      %s\n", orNotSet(cfg.Model))
	fmt.Printf("API key    %s\n", describeKey(cfg.APIKey))
	fmt.Println()

	if err := cfg.validate(); err != nil {
		printCheck("✗", "Config incomplete", err.Error())
		return 1
	}

	checkModels(cfg)

	start := time.Now()
	_, err := chatCompletion(cfg, chatRequest{
		Model:     cfg.Model,
		Messages:  []chatMessage{{Role: "user", Content: "Reply with OK"}},
		MaxTokens: 5,
	}, requestTimeout)
	if err != nil {
		printCheck("✗", "Generation failed", err.Error())
		return 1
	}
	printCheck("✓", "Generation works", latency(start))
	return 0
}

// checkModels reports server reachability and model availability. Failures
// are warnings only.
func checkModels(cfg config) {
	start := time.Now()
	data, err := apiRequest(cfg, http.MethodGet, "/models", nil, modelsTimeout)
	var statusErr *httpStatusError
	switch {
	case errors.As(err, &statusErr) && (statusErr.Code == http.StatusUnauthorized || statusErr.Code == http.StatusForbidden):
		printCheck("⚠", "API key rejected", fmt.Sprintf("/models returned %d", statusErr.Code))
		return
	case err != nil:
		printCheck("⚠", "Server check failed", err.Error())
		return
	}
	printCheck("✓", "Server reachable", latency(start))

	var models struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(data, &models); err != nil {
		printCheck("⚠", "Model list unreadable", err.Error())
		return
	}
	for _, m := range models.Data {
		if m.ID == cfg.Model {
			printCheck("✓", "Model available", "")
			return
		}
	}
	printCheck("⚠", "Model not listed", cfg.Model+" is not in /models")
}

func printCheck(mark, label, detail string) {
	if detail == "" {
		fmt.Printf("%s %s\n", mark, label)
		return
	}
	fmt.Printf("%s %-24s%s\n", mark, label, detail)
}

func latency(start time.Time) string {
	return fmt.Sprintf("%dms", time.Since(start).Milliseconds())
}

func orNotSet(v string) string {
	if v == "" {
		return "not set"
	}
	return v
}

// describeKey masks the API key, showing only its prefix and last 4 chars.
func describeKey(key string) string {
	if key == "" {
		return "not set"
	}
	if len(key) < 12 {
		return "set"
	}
	return fmt.Sprintf("set (%s…%s)", key[:3], key[len(key)-4:])
}
