package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSuggestSubjectsEmptyReply(t *testing.T) {
	tests := []struct {
		name         string
		finishReason string
		want         error
	}{
		{"stop", "stop", errNoSubjects},
		{"token limit", "length", errTokenLimit},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":""},"finish_reason":"` + tt.finishReason + `"}]}`))
			}))
			defer srv.Close()

			_, err := suggestSubjects(config{BaseURL: srv.URL, Model: "m"}, "diff", 2, nil)
			if err != tt.want {
				t.Fatalf("err = %v, want %v", err, tt.want)
			}
			if !errors.Is(err, errNoSubjects) {
				t.Fatalf("err = %v, want it to wrap errNoSubjects", err)
			}
		})
	}
}

func TestChatCompletionDropsRejectedParams(t *testing.T) {
	var bodies []map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		bodies = append(bodies, body)
		switch {
		case body["max_tokens"] != nil:
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"error":{"message":"Unsupported parameter: 'max_tokens' is not supported with this model."}}`))
		case body["temperature"] != nil:
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"error":{"message":"Unsupported value: 'temperature' does not support 0 with this model."}}`))
		default:
			_, _ = w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":"Add parser"},"finish_reason":"stop"}]}`))
		}
	}))
	defer srv.Close()

	subjects, err := suggestSubjects(config{BaseURL: srv.URL, Model: "m"}, "diff", 1, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(subjects) != 1 || subjects[0] != "Add parser" {
		t.Fatalf("subjects = %q", subjects)
	}
	if len(bodies) != 3 {
		t.Fatalf("requests = %d, want 3", len(bodies))
	}
	if bodies[2]["max_completion_tokens"] != float64(maxResponseTokens) {
		t.Errorf("max_completion_tokens = %v, want %d", bodies[2]["max_completion_tokens"], maxResponseTokens)
	}
}

func TestChatCompletionKeepsOtherErrors(t *testing.T) {
	requests := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"bad model"}`))
	}))
	defer srv.Close()

	if _, err := suggestSubjects(config{BaseURL: srv.URL, Model: "m"}, "diff", 1, nil); err == nil {
		t.Fatal("want error")
	}
	if requests != 1 {
		t.Fatalf("requests = %d, want 1", requests)
	}
}
