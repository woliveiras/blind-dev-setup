//go:build !windows

package target

import "errors"

func Inspect(path string) (Details, error) {
	return Details{}, errors.New("a preparação só pode ser executada no Windows")
}
