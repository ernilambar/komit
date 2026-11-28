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
	maxDiffBytes        = 4 * 1024
	diffTruncatedMarker = "\n[diff truncated]\n"
	// fileTruncatedMarker follows a file section whose hunks were cut short.
	fileTruncatedMarker = "\n[file diff truncated]\n"
)

// diffPathspecs cover the whole repo even when komit runs from a subdirectory.
var diffPathspecs = []string{
	":/",
	":(top,exclude)*.lock",
	":(top,exclude)package-lock.json",
	":(top,exclude)go.sum",
	":(top,exclude)*.min.*",
	":(top,exclude)*.map",
	":(top,exclude)*.svg",
	":(top,exclude,glob)**/dist/**",
	":(top,exclude,glob)**/vendor/**",
	":(top,exclude,glob)**/__snapshots__/**",
}

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

// stagedFiles lists staged paths with their A/M/D/R status, which tells the
// model directly whether a file was added, changed, removed or renamed.
func stagedFiles() (string, error) {
	out, err := gitOutput("diff", "--cached", "--name-status")
	if err != nil {
		return "", fmt.Errorf("git diff --cached --name-status: %w", err)
	}
	return out, nil
}

func stagedDiff() (string, error) {
	// One context line leaves more of the budget for the changes themselves.
	args := append([]string{"diff", "--cached", "--no-ext-diff", "-U1", "--"}, diffPathspecs...)
	out, err := gitOutput(args...)
	if err != nil {
		return "", fmt.Errorf("git diff --cached: %w", err)
	}
	return capDiff(out), nil
}

// capDiff keeps out within maxDiffBytes while giving every file its header and
// an equal share of hunk content, so a large early file cannot hide the files
// after it. When even the headers do not fit, out is cut as before.
func capDiff(out string) string {
	if len(out) <= maxDiffBytes {
		return out
	}

	preamble, sections := splitDiffSections(out)
	headers := make([]string, len(sections))
	bodies := make([]string, len(sections))
	fixed := len(preamble)
	for i, section := range sections {
		headers[i], bodies[i] = splitDiffSection(section)
		fixed += len(headers[i])
	}
	if len(sections) == 0 || fixed >= maxDiffBytes {
		return truncateUTF8(out, maxDiffBytes) + diffTruncatedMarker
	}

	quota := allocateDiff(len(sections), bodies, maxDiffBytes-fixed)

	var b strings.Builder
	b.Grow(maxDiffBytes + len(fileTruncatedMarker)*len(sections))
	b.WriteString(preamble)
	for i := range sections {
		b.WriteString(headers[i])
		if len(bodies[i]) > quota[i] {
			b.WriteString(cutLines(bodies[i], quota[i]))
			b.WriteString(fileTruncatedMarker)
			continue
		}
		b.WriteString(bodies[i])
	}
	return b.String()
}

// splitDiffSections splits diff output into the text before the first file and
// one section per file, each starting at its "diff --git" line.
func splitDiffSections(out string) (preamble string, sections []string) {
	var starts []int
	for off := 0; off < len(out); {
		end := strings.IndexByte(out[off:], '\n')
		if end < 0 {
			end = len(out)
		} else {
			end += off + 1
		}
		if strings.HasPrefix(out[off:], "diff --git ") {
			starts = append(starts, off)
		}
		off = end
	}
	if len(starts) == 0 {
		return out, nil
	}
	preamble = out[:starts[0]]
	for i, start := range starts {
		end := len(out)
		if i+1 < len(starts) {
			end = starts[i+1]
		}
		sections = append(sections, out[start:end])
	}
	return preamble, sections
}

// splitDiffSection splits a file section into its header and its hunks. Binary
// payloads count as body so they can be cut.
func splitDiffSection(section string) (header, body string) {
	for off := 0; off < len(section); {
		end := strings.IndexByte(section[off:], '\n')
		var line string
		if end < 0 {
			line = section[off:]
		} else {
			line = section[off : off+end+1]
		}
		if strings.HasPrefix(line, "@@") || strings.HasPrefix(line, "GIT binary patch") {
			return section[:off], section[off:]
		}
		if end < 0 {
			break
		}
		off += end + 1
	}
	return section, ""
}

// allocateDiff shares budget bytes between bodies, splitting it evenly and
// handing what shorter files leave unused to the longer ones.
func allocateDiff(n int, bodies []string, budget int) []int {
	quota := make([]int, n)
	active := make([]int, 0, n)
	for i := range bodies {
		active = append(active, i)
	}
	for len(active) > 0 {
		share := budget / len(active)
		if share == 0 {
			break
		}
		next := active[:0]
		for _, i := range active {
			if len(bodies[i]) <= share {
				quota[i] = len(bodies[i])
				budget -= len(bodies[i])
				continue
			}
			next = append(next, i)
		}
		if len(next) == len(active) {
			for _, i := range next {
				quota[i] = share
			}
			break
		}
		active = next
	}
	return quota
}

// cutLines returns the longest prefix of s ending on a line boundary that is at
// most limit bytes long.
func cutLines(s string, limit int) string {
	if limit >= len(s) {
		return s
	}
	if limit <= 0 {
		return ""
	}
	if i := strings.LastIndexByte(s[:limit], '\n'); i >= 0 {
		return s[:i+1]
	}
	return ""
}

// buildModelInput assembles the staged file list and the filtered, size-capped
// diff sent to the model.
func buildModelInput() (string, error) {
	files, err := stagedFiles()
	if err != nil {
		return "", err
	}
	diff, err := stagedDiff()
	if err != nil {
		return "", err
	}

	var b strings.Builder
	b.WriteString("Staged files (A added, M modified, D deleted, R renamed):\n")
	b.WriteString(files)
	if files != "" && !strings.HasSuffix(files, "\n") {
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
