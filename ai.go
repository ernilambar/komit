package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"syscall"
	"time"
)

const (
	requestTimeout = 60 * time.Second
	maxErrorBody   = 200
)

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model    string        `json:"model"`
	Messages []chatMessage `json:"messages"`
}

type chatResponse struct {
	Choices []struct {
		Message chatMessage `json:"message"`
	} `json:"choices"`
}

func systemPrompt(n int) string {
	return fmt.Sprintf(`You write Git commit subjects for a staged diff.

Rules:
- Plain imperative mood, e.g. "Add endpoint validation", "Fix nil pointer in parser".
- Single line, at most %d characters.
- No Conventional Commit prefixes (no "feat:", "fix:", etc.).
- No trailing period, no body, no trailers.

Return exactly %d subjects, one per line, nothing else.`, maxSubjectLen, n)
}

// generateSubjects asks the model for n commit subjects and returns its raw
// reply for cleanup.
func generateSubjects(cfg config, input string, n int) (string, error) {
	body, err := json.Marshal(chatRequest{
		Model: cfg.Model,
		Messages: []chatMessage{
			{Role: "system", Content: systemPrompt(n)},
			{Role: "user", Content: input},
		},
	})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest(http.MethodPost, cfg.BaseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	if cfg.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+cfg.APIKey)
	}

	client := &http.Client{Timeout: requestTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return "", describeRequestError(err, cfg.BaseURL)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", describeRequestError(err, cfg.BaseURL)
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return "", fmt.Errorf("server returned %d: %s", resp.StatusCode, shortBody(data))
	}

	var parsed chatResponse
	if err := json.Unmarshal(data, &parsed); err != nil {
		return "", fmt.Errorf("invalid response from server: %w", err)
	}
	if len(parsed.Choices) == 0 {
		return "", errors.New("server returned no choices")
	}
	return parsed.Choices[0].Message.Content, nil
}

func describeRequestError(err error, baseURL string) error {
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return fmt.Errorf("request to %s timed out after %s", baseURL, requestTimeout)
	}
	var dnsErr *net.DNSError
	if errors.Is(err, syscall.ECONNREFUSED) || errors.As(err, &dnsErr) {
		return fmt.Errorf("can't reach %s", baseURL)
	}
	return fmt.Errorf("request to %s failed: %w", baseURL, err)
}

func shortBody(data []byte) string {
	s := strings.Join(strings.Fields(string(data)), " ")
	if s == "" {
		return "(empty body)"
	}
	if len(s) > maxErrorBody {
		s = truncateUTF8(s, maxErrorBody) + "…"
	}
	return s
}
