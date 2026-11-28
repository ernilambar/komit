package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// fakeDiff builds git diff output with files sections, each carrying lines
// body lines after its hunk header.
func fakeDiff(files, body int) string {
	var b strings.Builder
	for i := range files {
		fmt.Fprintf(&b, "diff --git a/f%d.txt b/f%d.txt\n", i, i)
		fmt.Fprintf(&b, "index 0000000..1111111 100644\n--- a/f%d.txt\n+++ b/f%d.txt\n", i, i)
		fmt.Fprintf(&b, "@@ -1,%d +1,%d @@\n", body, body)
		for j := range body {
			fmt.Fprintf(&b, "+change %d in file %d with padding\n", j, i)
		}
	}
	return b.String()
}

func TestCapDiffKeepsEveryFileHeader(t *testing.T) {
	const files = 4
	out := fakeDiff(files, 200)
	if len(out) <= maxDiffBytes {
		t.Fatalf("fixture is %d bytes, want more than %d", len(out), maxDiffBytes)
	}

	got := capDiff(out)
	if n := strings.Count(got, "diff --git "); n != files {
		t.Errorf("kept %d file headers, want %d", n, files)
	}
	if n := strings.Count(got, fileTruncatedMarker); n != files {
		t.Errorf("marked %d truncated files, want %d", n, files)
	}
	if limit := maxDiffBytes + files*len(fileTruncatedMarker); len(got) > limit {
		t.Errorf("capDiff returned %d bytes, want at most %d", len(got), limit)
	}
	if !strings.HasSuffix(got, "\n") {
		t.Error("capDiff cut mid-line")
	}
	last := -1
	for i := range files {
		idx := strings.Index(got, fmt.Sprintf("+++ b/f%d.txt", i))
		if idx < 0 {
			t.Fatalf("file %d header missing", i)
		}
		if idx < last {
			t.Errorf("file %d out of order", i)
		}
		last = idx
		if !strings.Contains(got, fmt.Sprintf("+change 0 in file %d", i)) {
			t.Errorf("file %d lost its first body line", i)
		}
	}
}

func TestCapDiffEqualSharesLetShortFilesGiveBack(t *testing.T) {
	const body = 300
	small := "diff --git a/small.txt b/small.txt\n--- a/small.txt\n+++ b/small.txt\n@@ -1 +1 @@\n+tiny\n"
	var b strings.Builder
	b.WriteString(small)
	for _, name := range []string{"big1.txt", "big2.txt"} {
		fmt.Fprintf(&b, "diff --git a/%s b/%s\n--- a/%s\n+++ b/%s\n@@ -1,%d +1,%d @@\n", name, name, name, name, body, body)
		for j := range body {
			fmt.Fprintf(&b, "+%s line %d with padding padding padding\n", name, j)
		}
	}

	got := capDiff(b.String())
	if !strings.Contains(got, small) {
		t.Error("small file was cut or marked truncated")
	}
	if n := strings.Count(got, fileTruncatedMarker); n != 2 {
		t.Errorf("marked %d files truncated, want 2", n)
	}
}

func TestAllocateDiffGivesBackUnusedShare(t *testing.T) {
	bodies := []string{strings.Repeat("x", 10), strings.Repeat("x", 1000), strings.Repeat("x", 1000)}
	got := allocateDiff(len(bodies), bodies, 900)
	want := []int{10, 445, 445}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("allocateDiff = %v, want %v", got, want)
		}
	}
}

func TestCapDiffFallsBackWhenHeadersDoNotFit(t *testing.T) {
	out := fakeDiff(300, 1)
	got := capDiff(out)
	if len(got) > maxDiffBytes+len(diffTruncatedMarker) {
		t.Errorf("capDiff returned %d bytes, want at most %d", len(got), maxDiffBytes+len(diffTruncatedMarker))
	}
	if !strings.Contains(got, diffTruncatedMarker) {
		t.Error("fallback output is missing the truncation marker")
	}
}

func TestCapFiles(t *testing.T) {
	list := func(n int) string {
		var b strings.Builder
		for i := range n {
			fmt.Fprintf(&b, "M\tfile%d.go\n", i)
		}
		return b.String()
	}
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"empty", "", ""},
		{"under cap", list(3), list(3)},
		{"at cap", list(maxFileLines), list(maxFileLines)},
		{"over cap", list(maxFileLines + 7), list(maxFileLines) + "… and 7 more files\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := capFiles(tt.in); got != tt.want {
				t.Fatalf("capFiles() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCapDiffKeepsSmallDiff(t *testing.T) {
	out := fakeDiff(2, 3)
	if got := capDiff(out); got != out {
		t.Errorf("capDiff changed a diff under the cap:\n got: %q\nwant: %q", got, out)
	}
}

func TestStagedDiffFromSubdirectory(t *testing.T) {
	dir := t.TempDir()
	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	write := func(name, content string) {
		t.Helper()
		path := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	run("init", "-q")
	write("root.txt", "root\n")
	write("sub/s.txt", "sub\n")
	write("sub/deps.lock", "lock\n")
	run("add", ".")
	t.Chdir(filepath.Join(dir, "sub"))

	diff, err := stagedDiff()
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"root.txt", "sub/s.txt"} {
		if !strings.Contains(diff, want) {
			t.Errorf("diff is missing %s", want)
		}
	}
	if strings.Contains(diff, "deps.lock") {
		t.Error("diff includes excluded deps.lock")
	}
}
