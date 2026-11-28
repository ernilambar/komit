package main

import (
	"fmt"
	"os"
	"strings"
)

type config struct {
	BaseURL string
	Model   string
	APIKey  string
	// ReasoningEffort is sent as reasoning_effort when set, e.g. "none" to
	// skip thinking on reasoning models.
	ReasoningEffort string
}

func loadConfig() config {
	return config{
		BaseURL:         strings.TrimRight(os.Getenv("KOMIT_BASE_URL"), "/"),
		Model:           os.Getenv("KOMIT_MODEL"),
		APIKey:          os.Getenv("KOMIT_API_KEY"),
		ReasoningEffort: os.Getenv("KOMIT_REASONING_EFFORT"),
	}
}

// validate reports all missing required settings in one error.
func (c config) validate() error {
	var missing []string
	if c.BaseURL == "" {
		missing = append(missing, "KOMIT_BASE_URL")
	}
	if c.Model == "" {
		missing = append(missing, "KOMIT_MODEL")
	}
	if len(missing) == 0 {
		return nil
	}
	return fmt.Errorf("%s not set", strings.Join(missing, " and "))
}
