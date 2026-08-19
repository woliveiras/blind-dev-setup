//go:build windows

package target

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"unsafe"
)

var (
	kernel32              = syscall.NewLazyDLL("kernel32.dll")
	getVolumeInformationW = kernel32.NewProc("GetVolumeInformationW")
	getDiskFreeSpaceExW   = kernel32.NewProc("GetDiskFreeSpaceExW")
)

func Inspect(path string) (Details, error) {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return Details{}, fmt.Errorf("resolver destino: %w", err)
	}
	info, err := os.Stat(absolute)
	if err != nil {
		return Details{}, fmt.Errorf("abrir destino: %w", err)
	}
	if !info.IsDir() {
		return Details{}, fmt.Errorf("o destino não é um diretório: %s", absolute)
	}
	volume := filepath.VolumeName(absolute)
	if volume == "" {
		return Details{}, fmt.Errorf("não foi possível identificar o volume de %s", absolute)
	}
	root := volume + "\\"
	fileSystem, err := volumeFileSystem(root)
	if err != nil {
		return Details{}, err
	}
	freeBytes, err := volumeFreeBytes(root)
	if err != nil {
		return Details{}, err
	}
	_, destinationErr := os.Stat(filepath.Join(absolute, DirectoryName))
	if destinationErr != nil && !os.IsNotExist(destinationErr) {
		return Details{}, fmt.Errorf("verificar instalação existente: %w", destinationErr)
	}

	return Details{
		Path:              absolute,
		Volume:            volume,
		SystemVolume:      filepath.VolumeName(os.Getenv("SystemDrive") + "\\"),
		FileSystem:        fileSystem,
		FreeBytes:         freeBytes,
		DestinationExists: destinationErr == nil,
	}, nil
}

func volumeFileSystem(root string) (string, error) {
	rootPointer, err := syscall.UTF16PtrFromString(root)
	if err != nil {
		return "", fmt.Errorf("resolver volume: %w", err)
	}
	fileSystemName := make([]uint16, 32)
	result, _, callErr := getVolumeInformationW.Call(
		uintptr(unsafe.Pointer(rootPointer)),
		0,
		0,
		0,
		0,
		0,
		uintptr(unsafe.Pointer(&fileSystemName[0])),
		uintptr(len(fileSystemName)),
	)
	if result == 0 {
		return "", fmt.Errorf("consultar sistema de arquivos: %w", callErr)
	}
	return syscall.UTF16ToString(fileSystemName), nil
}

func volumeFreeBytes(root string) (uint64, error) {
	rootPointer, err := syscall.UTF16PtrFromString(root)
	if err != nil {
		return 0, fmt.Errorf("resolver volume: %w", err)
	}
	var available uint64
	result, _, callErr := getDiskFreeSpaceExW.Call(
		uintptr(unsafe.Pointer(rootPointer)),
		uintptr(unsafe.Pointer(&available)),
		0,
		0,
	)
	if result == 0 {
		return 0, fmt.Errorf("consultar espaço livre: %w", callErr)
	}
	return available, nil
}
