package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

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
