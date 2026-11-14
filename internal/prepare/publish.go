package prepare

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"
	"syscall"
	"time"
)

// Windows can temporarily deny a directory rename while recently executed
// binaries or antivirus scanners still hold handles in that directory.
// Keep publication a rename: copying would expose a partially published bundle.
func publish(ctx context.Context, staging, destination string, output io.Writer) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	announced := false
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		// A destination created during preparation must never be replaced.
		if _, err := os.Lstat(destination); err == nil {
			return fmt.Errorf("o destino já existe: %w", os.ErrExist)
		} else if !errors.Is(err, os.ErrNotExist) {
			return err
		}
		err := os.Rename(staging, destination)
		if err == nil || !retryablePublishError(err) {
			return err
		}
		if !announced {
			fmt.Fprintln(output, "O Windows está usando arquivos do ambiente; aguardando até 30 segundos para publicar.")
			announced = true
		}
		timer := time.NewTimer(200 * time.Millisecond)
		select {
		case <-ctx.Done():
			timer.Stop()
			return errors.Join(err, ctx.Err())
		case <-timer.C:
		}
	}
}

func retryablePublishError(err error) bool {
	if runtime.GOOS != "windows" {
		return false
	}
	// Win32 ERROR_ACCESS_DENIED, ERROR_SHARING_VIOLATION, ERROR_LOCK_VIOLATION.
	return errors.Is(err, syscall.Errno(5)) ||
		errors.Is(err, syscall.Errno(32)) ||
		errors.Is(err, syscall.Errno(33))
}
