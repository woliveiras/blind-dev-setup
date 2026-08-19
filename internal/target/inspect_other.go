//go:build !windows

package target

import "errors"

func Inspect(path string) (Details, error) {
	return Details{}, errors.New("a preparação da v0.1.0 só pode ser executada no Windows")
}
