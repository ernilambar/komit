package main

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

const maxSubjectLen = 72

var (
	thinkBlockRe   = regexp.MustCompile(`(?is)<think\b[^>]*>.*?</think[^>]*>`)
	thinkOpenRe    = regexp.MustCompile(`(?i)<think\b`)
	thinkCloseRe   = regexp.MustCompile(`(?i)</think[^>]*>`)
	conventionalRe = regexp.MustCompile(`(?i)^(feat|fix|chore|docs|style|refactor|perf|test|build|ci|revert)(\([^)]*\))?!?:\s*`)
	numberingRe    = regexp.MustCompile(`^\d+[.)]\s*`)
	bulletRe       = regexp.MustCompile(`^[-*](?:\s+|$)`)
	headingRe      = regexp.MustCompile(`^#{1,6}(?:\s+|$)`)
)

// cleanSubjects turns raw model output into commit subjects that satisfy the
// style rules. Lines that cannot be salvaged are dropped.
func cleanSubjects(raw string) []string {
	raw = stripThinking(raw)
	seen := make(map[string]struct{})
	var subjects []string
	for _, line := range strings.Split(raw, "\n") {
		subject, ok := cleanLine(line)
		if !ok {
			continue
		}
		key := strings.ToLower(subject)
		if _, dup := seen[key]; dup {
			continue
		}
		seen[key] = struct{}{}
		subjects = append(subjects, subject)
	}
	return subjects
}

// stripThinking removes reasoning blocks. A stray closing tag drops everything
// before it; an unclosed opening tag drops everything after it.
func stripThinking(raw string) string {
	raw = thinkBlockRe.ReplaceAllString(raw, "")
	if locs := thinkCloseRe.FindAllStringIndex(raw, -1); len(locs) > 0 {
		raw = raw[locs[len(locs)-1][1]:]
	}
	if loc := thinkOpenRe.FindStringIndex(raw); loc != nil {
		raw = raw[:loc[0]]
	}
	return raw
}

// cleanLine cleans one line of model output and reports whether it is a valid
// subject. Lines ending in a colon are preamble, not subjects.
func cleanLine(line string) (string, bool) {
	line = stripDecorations(line)
	line = conventionalRe.ReplaceAllString(line, "")
	line = strings.TrimSpace(line)
	line = strings.TrimSpace(strings.TrimRight(line, "."))
	if line == "" || strings.HasSuffix(line, ":") || utf8.RuneCountInString(line) > maxSubjectLen {
		return "", false
	}
	return capitalize(line), true
}

var errSubjectLength = fmt.Errorf("subject must be 1-%d characters", maxSubjectLen)

// cleanInput validates a user-typed subject. Unlike model output, a
// Conventional Commit prefix is rejected rather than silently stripped.
func cleanInput(line string) (string, error) {
	if m := conventionalRe.FindString(strings.TrimSpace(line)); m != "" {
		return "", fmt.Errorf("remove the %q prefix", strings.TrimSpace(m))
	}
	if strings.HasSuffix(strings.TrimSpace(line), ":") {
		return "", errors.New("subject must not end with a colon")
	}
	subject, ok := cleanLine(line)
	if !ok {
		return "", errSubjectLength
	}
	return subject, nil
}

func stripDecorations(line string) string {
	for {
		previous := line
		line = strings.TrimSpace(line)
		line = bulletRe.ReplaceAllString(line, "")
		line = numberingRe.ReplaceAllString(line, "")
		line = headingRe.ReplaceAllString(line, "")
		line = strings.Trim(line, "\"'`*")
		if line == previous {
			return line
		}
	}
}

func capitalize(s string) string {
	r, size := utf8.DecodeRuneInString(s)
	if r == utf8.RuneError && size <= 1 {
		return s
	}
	return string(unicode.ToUpper(r)) + s[size:]
}
