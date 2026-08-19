//go:build windows

package target

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const listTargetsPowerShell = `$ErrorActionPreference = 'Stop'
[Console]::OutputEncoding = [System.Text.UTF8Encoding]::new($false)
$items = @()
$disks = @(Get-Disk -ErrorAction Stop | Where-Object { $_.BusType -eq 'USB' })
foreach ($disk in $disks) {
    $volumes = @()
    $partitions = @($disk | Get-Partition -ErrorAction SilentlyContinue)
    foreach ($partition in $partitions) {
        $volume = $partition | Get-Volume -ErrorAction SilentlyContinue
        if ($null -ne $volume) {
            $volumes += @($volume)
        }
    }
    if ($volumes.Count -eq 0) {
        $items += [pscustomobject]@{
            DiskNumber = [uint32]$disk.Number
            Model = [string]$disk.FriendlyName
            DriveLetter = ''
            Label = ''
            FileSystem = ''
            SizeBytes = [uint64]$disk.Size
            FreeBytes = [uint64]0
        }
        continue
    }
    foreach ($volume in $volumes) {
        $items += [pscustomobject]@{
            DiskNumber = [uint32]$disk.Number
            Model = [string]$disk.FriendlyName
            DriveLetter = [string]$volume.DriveLetter
            Label = [string]$volume.FileSystemLabel
            FileSystem = [string]$volume.FileSystem
            SizeBytes = [uint64]$volume.Size
            FreeBytes = [uint64]$volume.SizeRemaining
        }
    }
}
ConvertTo-Json -InputObject @($items) -Compress -Depth 3`

func List() ([]Candidate, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	command := exec.CommandContext(
		ctx,
		"powershell.exe",
		"-NoLogo",
		"-NoProfile",
		"-NonInteractive",
		"-Command",
		listTargetsPowerShell,
	)
	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	if err := command.Run(); err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return nil, errors.New("o Windows demorou demais para responder")
		}
		details := strings.TrimSpace(stderr.String())
		if details == "" {
			return nil, fmt.Errorf("consultar dispositivos USB: %w", err)
		}
		return nil, fmt.Errorf("consultar dispositivos USB: %s", details)
	}

	candidates, err := decodeCandidates(stdout.Bytes())
	if err != nil {
		return nil, err
	}
	if len(candidates) == 0 {
		return candidates, nil
	}
	systemVolume, err := windowsSystemVolume()
	if err != nil {
		return nil, err
	}
	for index := range candidates {
		root := candidates[index].RootPath()
		if root == "" {
			continue
		}
		candidates[index].IsSystem = strings.EqualFold(
			strings.TrimSuffix(root, "\\"),
			systemVolume,
		)
		_, statErr := os.Stat(filepath.Join(root, DirectoryName))
		candidates[index].DestinationExists = statErr == nil
	}
	return candidates, nil
}
