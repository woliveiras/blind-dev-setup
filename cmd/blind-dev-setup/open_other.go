//go:build !windows

package main

import "errors"

func openEnvironment(string) error {
	return errors.New("abrir a pasta do ambiente só está disponível no Windows")
}
