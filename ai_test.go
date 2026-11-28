package main

import (
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

			_, err := suggestSubjects(config{BaseURL: srv.URL, Model: "m"}, "diff", 2)
			if err != tt.want {
				t.Fatalf("err = %v, want %v", err, tt.want)
			}
			if !errors.Is(err, errNoSubjects) {
				t.Fatalf("err = %v, want it to wrap errNoSubjects", err)
			}
		})
	}
}
