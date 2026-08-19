package target

import (
	"errors"
	"fmt"
	"strings"
)

const DirectoryName = "blind-dev-setup"

type Details struct {
	Path              string
	Volume            string
	SystemVolume      string
	FileSystem        string
	FreeBytes         uint64
	DestinationExists bool
}

func Validate(details Details, minimumFreeBytes int64) error {
	if details.Path == "" || details.Volume == "" {
		return errors.New("não foi possível identificar o volume de destino")
	}
	if details.SystemVolume == "" {
		return errors.New("não foi possível identificar o volume do sistema")
	}
	if strings.EqualFold(details.Volume, details.SystemVolume) {
		return errors.New("o destino está no disco do sistema; escolha outro volume")
	}
	if !strings.EqualFold(details.FileSystem, "NTFS") {
		return fmt.Errorf("o destino usa %s; a v0.1.0 exige NTFS", details.FileSystem)
	}
	if details.DestinationExists {
		return fmt.Errorf("%s já existe no destino; a v0.1.0 não sobrescreve instalações", DirectoryName)
	}
	if minimumFreeBytes <= 0 {
		return errors.New("o requisito de espaço é inválido")
	}
	if details.FreeBytes < uint64(minimumFreeBytes) {
		return fmt.Errorf(
			"espaço livre insuficiente: %.1f GiB disponíveis, %.1f GiB necessários",
			float64(details.FreeBytes)/(1<<30),
			float64(minimumFreeBytes)/(1<<30),
		)
	}
	return nil
}
