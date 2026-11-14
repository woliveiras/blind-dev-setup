package prepare

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestPublishDoesNotReplaceExistingDestination(t *testing.T) {
	root := t.TempDir()
	staging := filepath.Join(root, "staging")
	destination := filepath.Join(root, "published")
	for _, path := range []string{staging, destination} {
		if err := os.Mkdir(path, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(staging, "READY.txt"), []byte("ready"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := publish(context.Background(), staging, destination, io.Discard); !errors.Is(err, os.ErrExist) {
		t.Fatalf("publish() error = %v, want existing destination error", err)
	}
	if _, err := os.Stat(filepath.Join(staging, "READY.txt")); err != nil {
		t.Fatalf("staging was changed: %v", err)
	}
	entries, err := os.ReadDir(destination)
	if err != nil || len(entries) != 0 {
		t.Fatalf("existing destination was changed: %v, %v", entries, err)
	}
}

func TestPublishReturnsPermanentError(t *testing.T) {
	root := t.TempDir()
	err := publish(context.Background(), filepath.Join(root, "missing"), filepath.Join(root, "published"), io.Discard)
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("publish() error = %v, want missing source error", err)
	}
}

func TestPublishHonorsCancellation(t *testing.T) {
	root := t.TempDir()
	staging := filepath.Join(root, "staging")
	if err := os.Mkdir(staging, 0o755); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := publish(ctx, staging, filepath.Join(root, "published"), io.Discard); !errors.Is(err, context.Canceled) {
		t.Fatalf("publish() error = %v, want cancellation", err)
	}
	if _, err := os.Stat(staging); err != nil {
		t.Fatalf("cancelled publication changed staging: %v", err)
	}
}
