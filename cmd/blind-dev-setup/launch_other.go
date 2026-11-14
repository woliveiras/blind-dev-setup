//go:build !windows

package main

import (
	"context"
	"errors"
	"io"
)

func launchEnvironment(context.Context, string, io.Writer) error {
	return errors.New("iniciar o ambiente só está disponível no Windows")
}
