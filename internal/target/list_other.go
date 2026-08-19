//go:build !windows

package target

import "errors"

func List() ([]Candidate, error) {
	return nil, errors.New("a procura de pendrives da v0.1.0 só está disponível no Windows")
}
