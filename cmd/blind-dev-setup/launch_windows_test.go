//go:build windows

package main

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLaunchEnvironmentRunsFixedScriptInSelectedDirectory(t *testing.T) {
	root := launchFixture(t, "@echo off\nif not \"%BLIND_DEV_NO_PAUSE%\"==\"1\" exit /b 10\nif not \"%~1\"==\"\" exit /b 11\necho \"%CD%\"\necho captured-stderr 1>&2\necho launched> launched.txt\npause >nul\nexit /b 0\n")
	// Neither PATH nor COMSPEC may replace the trusted Windows interpreter.
	t.Setenv("PATH", root)
	t.Setenv("ComSpec", filepath.Join(root, "untrusted-cmd.exe"))
	t.Setenv("UNCHANGED", "must-not-expand")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var output bytes.Buffer
	if err := launchEnvironment(ctx, root, &output); err != nil {
		t.Fatalf("error=%v output=%s", err, output.String())
	}
	printedRoot := strings.Trim(strings.SplitN(output.String(), "\n", 2)[0], "\"\r\n ")
	expectedInfo, err := os.Stat(root)
	if err != nil {
		t.Fatal(err)
	}
	actualInfo, err := os.Stat(printedRoot)
	// Windows can render a directory with either its long or short spelling.
	if err != nil || !os.SameFile(expectedInfo, actualInfo) || !strings.Contains(output.String(), "captured-stderr") {
		t.Fatalf("output=%s", output.String())
	}
	if _, err := os.Stat(filepath.Join(root, "launched.txt")); err != nil {
		t.Fatal(err)
	}
}

func TestLaunchEnvironmentReturnsInitializerErrors(t *testing.T) {
	for _, body := range []string{
		"@echo off\necho initializer-failed 1>&2\nexit /b 23\n",
		"@echo off\ncall .\\missing-initializer.cmd\nif errorlevel 1 exit /b 17\nexit /b 0\n",
	} {
		root := launchFixture(t, body)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		var output bytes.Buffer
		err := launchEnvironment(ctx, root, &output)
		cancel()
		if err == nil {
			t.Fatalf("initializer failure reported success: %s", output.String())
		}
		if !strings.Contains(err.Error(), "executar START.cmd") {
			t.Fatalf("error=%v", err)
		}
	}
}

func TestLaunchEnvironmentContextStopsBlockedInitializer(t *testing.T) {
	root := launchFixture(t, "@echo off\n:wait\ngoto wait\n")
	ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
	defer cancel()
	started := time.Now()
	err := launchEnvironment(ctx, root, io.Discard)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("error=%v", err)
	}
	if elapsed := time.Since(started); elapsed > 4*time.Second {
		t.Fatalf("cancel took %s", elapsed)
	}
}

func TestLaunchEnvironmentCanWriteDirectlyToLogFile(t *testing.T) {
	root := launchFixture(t, "@echo off\necho launcher-diagnostic\nexit /b 0\n")
	file, err := os.CreateTemp(t.TempDir(), "launch-*.log")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	if err := launchEnvironment(context.Background(), root, file); err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(file.Name())
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(content, []byte("launcher-diagnostic")) {
		t.Fatalf("log=%s", content)
	}
}
