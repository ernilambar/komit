package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"unicode/utf8"
)

const (
	maxDiffBytes        = 12 * 1024
	diffTruncatedMarker = "\n[diff truncated]\n"
)

var diffExcludePaths = []string{":!*.lock", ":!package-lock.json", ":!go.sum", ":!*.min.*"}

var (
	errNotRepo       = errors.New("not a git repository")
	errNothingStaged = errors.New("nothing staged — stage changes or pass -a")
)

// gitOutput runs a git command and returns its stdout. stderr is included in
// the error when git exits non-zero.
func gitOutput(args ...string) (string, error) {
	var stdout, stderr bytes.Buffer
	cmd := exec.Command("git", args...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return stdout.String(), errors.New(msg)
	}
	return stdout.String(), nil
}

func ensureRepo() error {
	out, err := gitOutput("rev-parse", "--is-inside-work-tree")
	if err != nil || strings.TrimSpace(out) != "true" {
		return errNotRepo
	}
	return nil
}

func stageAll() error {
	if _, err := gitOutput("add", "-A"); err != nil {
		return fmt.Errorf("git add -A: %w", err)
	}
	return nil
}

func checkStaged() error {
	var stderr bytes.Buffer
	cmd := exec.Command("git", "diff", "--cached", "--quiet")
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err == nil {
		return errNothingStaged
	}
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		return fmt.Errorf("git diff --cached --quiet: %w", err)
	}
	if exitErr.ExitCode() == 1 {
		return nil
	}
	msg := strings.TrimSpace(stderr.String())
	if msg == "" {
		msg = err.Error()
	}
	return fmt.Errorf("git diff --cached --quiet: %s", msg)
}

func stagedStat() (string, error) {
	out, err := gitOutput("diff", "--cached", "--stat")
	if err != nil {
		return "", fmt.Errorf("git diff --cached --stat: %w", err)
	}
	return out, nil
}

func stagedDiff() (string, error) {
	args := append([]string{"diff", "--cached", "--no-ext-diff", "--", "."}, diffExcludePaths...)
	out, err := gitOutput(args...)
	if err != nil {
		return "", fmt.Errorf("git diff --cached: %w", err)
	}
	if len(out) <= maxDiffBytes {
		return out, nil
	}
	return truncateUTF8(out, maxDiffBytes) + diffTruncatedMarker, nil
}

// buildModelInput assembles the staged stat and the filtered, size-capped diff
// sent to the model.
func buildModelInput() (string, error) {
	stat, err := stagedStat()
	if err != nil {
		return "", err
	}
	diff, err := stagedDiff()
	if err != nil {
		return "", err
	}

	var b strings.Builder
	b.WriteString("Staged changes (stat):\n")
	b.WriteString(stat)
	if stat != "" && !strings.HasSuffix(stat, "\n") {
		b.WriteByte('\n')
	}
	b.WriteString("\nStaged diff:\n")
	b.WriteString(diff)
	return b.String(), nil
}

// gitStream runs a git command with its output streamed to the terminal.
func gitStream(args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func commit(subject string) error {
	if err := gitStream("commit", "-m", subject); err != nil {
		return fmt.Errorf("git commit failed: %w", err)
	}
	return nil
}

// push pushes the current branch, setting origin as upstream when none exists.
func push() error {
	args := []string{"push"}
	if _, err := gitOutput("rev-parse", "--abbrev-ref", "@{u}"); err != nil {
		args = []string{"push", "-u", "origin", "HEAD"}
	}
	if err := gitStream(args...); err != nil {
		return fmt.Errorf("git %s failed (commit kept): %w", strings.Join(args, " "), err)
	}
	return nil
}

// truncateUTF8 cuts s to at most max bytes without splitting a rune.
func truncateUTF8(s string, max int) string {
	if len(s) <= max {
		return s
	}
	cut := max
	for cut > 0 && !utf8.RuneStart(s[cut]) {
		cut--
	}
	return s[:cut]
}
