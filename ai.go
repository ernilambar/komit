package main

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
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

// sessionID identifies this run to OpenCode, whose Go endpoint rejects
// requests without x-opencode-session.
var sessionID = newSessionID()

func isOpenCode(host string) bool {
	return host == "opencode.ai" || strings.HasSuffix(host, ".opencode.ai")
}

func newSessionID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return "komit-" + hex.EncodeToString(b)
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
	return chatCompletion(cfg, chatRequest{
		Model: cfg.Model,
		Messages: []chatMessage{
			{Role: "system", Content: systemPrompt(n)},
			{Role: "user", Content: input},
		},
	}, requestTimeout)
}

// httpStatusError reports a non-2xx response.
type httpStatusError struct {
	Code int
	Body string
}

func (e *httpStatusError) Error() string {
	return fmt.Sprintf("server returned %d: %s", e.Code, e.Body)
}

// apiRequest sends a request to {base_url}{path} and returns the body of a 2xx
// response. A nil payload sends no body.
func apiRequest(cfg config, method, path string, payload any, timeout time.Duration) ([]byte, error) {
	var body io.Reader
	if payload != nil {
		encoded, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		body = bytes.NewReader(encoded)
	}

	req, err := http.NewRequest(method, cfg.BaseURL+path, body)
	if err != nil {
		return nil, err
	}
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("User-Agent", "komit")
	if isOpenCode(req.URL.Hostname()) {
		req.Header.Set("x-opencode-session", sessionID)
	}
	if cfg.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+cfg.APIKey)
	}

	client := &http.Client{Timeout: timeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, describeRequestError(err, cfg.BaseURL, timeout)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, describeRequestError(err, cfg.BaseURL, timeout)
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, &httpStatusError{Code: resp.StatusCode, Body: shortBody(data)}
	}
	return data, nil
}

// chatCompletion posts to /chat/completions and returns the first choice.
func chatCompletion(cfg config, request chatRequest, timeout time.Duration) (string, error) {
	data, err := apiRequest(cfg, http.MethodPost, "/chat/completions", request, timeout)
	if err != nil {
		return "", err
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

var errNoSubjects = errors.New("model returned no valid subjects")

// suggestSubjects generates and cleans up to n subjects, regenerating once if
// every candidate is rejected.
func suggestSubjects(cfg config, input string, n int) ([]string, error) {
	for attempt := 0; attempt < 2; attempt++ {
		raw, err := generateSubjects(cfg, input, n)
		if err != nil {
			return nil, err
		}
		if subjects := cleanSubjects(raw); len(subjects) > 0 {
			if len(subjects) > n {
				subjects = subjects[:n]
			}
			return subjects, nil
		}
	}
	return nil, errNoSubjects
}

func describeRequestError(err error, baseURL string, timeout time.Duration) error {
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return fmt.Errorf("request to %s timed out after %s", baseURL, timeout)
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
