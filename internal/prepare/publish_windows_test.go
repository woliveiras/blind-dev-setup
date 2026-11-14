//go:build windows

package prepare

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

type publishRetryWriter func([]byte) (int, error)

func (w publishRetryWriter) Write(p []byte) (int, error) { return w(p) }

func lockedStaging(t *testing.T) (string, string, func()) {
	t.Helper()
	root := t.TempDir()
	staging := filepath.Join(root, "staging")
	if err := os.Mkdir(staging, 0o755); err != nil {
		t.Fatal(err)
	}
	path, err := syscall.UTF16PtrFromString(staging)
	if err != nil {
		t.Fatal(err)
	}
	handle, err := syscall.CreateFile(path, syscall.GENERIC_READ,
		syscall.FILE_SHARE_READ|syscall.FILE_SHARE_WRITE, nil,
		syscall.OPEN_EXISTING, syscall.FILE_FLAG_BACKUP_SEMANTICS, 0)
	if err != nil {
		t.Fatal(err)
	}
	closed := false
	unlock := func() {
		if !closed {
			if err := syscall.CloseHandle(handle); err != nil {
				t.Fatal(err)
			}
			closed = true
		}
	}
	t.Cleanup(unlock)
	return staging, filepath.Join(root, "published"), unlock
}

func TestPublishRetriesWindowsDirectoryLock(t *testing.T) {
	staging, destination, unlock := lockedStaging(t)
	if err := os.Rename(staging, destination); err == nil || !retryablePublishError(err) {
		t.Fatalf("expected a real Windows rename lock, got %v", err)
	}
	announcements := 0
	output := publishRetryWriter(func(p []byte) (int, error) {
		announcements++
		unlock()
		return len(p), nil
	})
	if err := publish(context.Background(), staging, destination, output); err != nil {
		t.Fatal(err)
	}
	if announcements != 1 {
		t.Fatalf("retry announcements = %d, want 1", announcements)
	}
	if _, err := os.Stat(destination); err != nil {
		t.Fatal(err)
	}
}

func TestPublishWindowsLockHonorsDeadline(t *testing.T) {
	staging, destination, _ := lockedStaging(t)
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	if err := publish(ctx, staging, destination, io.Discard); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("publish() error = %v, want deadline", err)
	}
	if _, err := os.Stat(destination); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("locked staging was published: %v", err)
	}
}
