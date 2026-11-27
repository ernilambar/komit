package main

import (
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

const maxSubjectLen = 72

var (
	thinkBlockRe   = regexp.MustCompile(`(?is)<think\b[^>]*>.*?</think[^>]*>`)
	conventionalRe = regexp.MustCompile(`(?i)^(feat|fix|chore|docs|style|refactor|perf|test|build|ci|revert)(\([^)]*\))?!?:\s*`)
	numberingRe    = regexp.MustCompile(`^\d+[.)]\s*`)
	bulletRe       = regexp.MustCompile(`^[-*](?:\s+|$)`)
)

// cleanSubjects turns raw model output into commit subjects that satisfy the
// style rules. Lines that cannot be salvaged are dropped.
func cleanSubjects(raw string) []string {
	raw = thinkBlockRe.ReplaceAllString(raw, "")
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

// cleanLine cleans one line and reports whether it is a valid subject. This is
// also the validation path for "Write my own" input.
func cleanLine(line string) (string, bool) {
	line = stripDecorations(line)
	line = conventionalRe.ReplaceAllString(line, "")
	line = strings.TrimSpace(line)
	line = strings.TrimSpace(strings.TrimRight(line, "."))
	if line == "" || utf8.RuneCountInString(line) > maxSubjectLen {
		return "", false
	}
	return capitalize(line), true
}

func stripDecorations(line string) string {
	for {
		previous := line
		line = strings.TrimSpace(line)
		line = bulletRe.ReplaceAllString(line, "")
		line = numberingRe.ReplaceAllString(line, "")
		line = strings.Trim(line, "\"'`")
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
