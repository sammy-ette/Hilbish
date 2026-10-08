//go:build !windows

package tests

import (
	"bytes"
	"context"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func repositoryRoot(t *testing.T) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("locating the tests package")
	}
	return filepath.Dir(filepath.Dir(filename))
}

func buildHilbish(t *testing.T) string {
	t.Helper()
	binary := filepath.Join(t.TempDir(), "hilbish")
	cmd := exec.Command("go", "build", "-ldflags=-checklinkname=0", "-o", binary, ".")
	cmd.Dir = repositoryRoot(t)
	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &output
	if err := cmd.Run(); err != nil {
		t.Fatalf("building Hilbish: %v\n%s", err, output.String())
	}
	return binary
}

func runHilbish(t *testing.T, binary, script string) (string, string) {
	t.Helper()
	return runHilbishWithEnv(t, binary, script, nil)
}

func runHilbishWithEnv(t *testing.T, binary, script string, env []string) (string, string) {
	t.Helper()
	return runHilbishArgsWithEnv(t, binary, "", env, "-c", script)
}

func runHilbishArgs(t *testing.T, binary, input string, args ...string) (string, string) {
	t.Helper()
	return runHilbishArgsWithEnv(t, binary, input, nil, args...)
}

func runHilbishArgsWithEnv(t *testing.T, binary, input string, env []string, args ...string) (string, string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, binary, args...)
	cmd.Dir = repositoryRoot(t)
	cmd.Stdin = strings.NewReader(input)
	if env != nil {
		cmd.Env = env
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		if ctx.Err() != nil {
			t.Fatalf("Hilbish command timed out")
		}
		t.Fatalf("Hilbish command failed: %v\nstdout: %q\nstderr: %q", err, stdout.String(), stderr.String())
	}
	return stdout.String(), stderr.String()
}

func outputLine(t *testing.T, output string) string {
	t.Helper()
	return strings.TrimSuffix(output, "\n")
}
