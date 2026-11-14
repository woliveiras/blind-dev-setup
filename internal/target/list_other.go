//go:build !windows

package target

import "errors"

func List() ([]Candidate, error) {
	return nil, errors.New("a procura de pendrives só está disponível no Windows")
}
