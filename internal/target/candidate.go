package target

import (
	"fmt"
	"strings"
)

type Candidate struct {
	DiskNumber        uint32 `json:"DiskNumber"`
	Model             string `json:"Model"`
	DriveLetter       string `json:"DriveLetter"`
	Label             string `json:"Label"`
	FileSystem        string `json:"FileSystem"`
	SizeBytes         uint64 `json:"SizeBytes"`
	FreeBytes         uint64 `json:"FreeBytes"`
	IsSystem          bool   `json:"-"`
	DestinationExists bool   `json:"-"`
}

func (candidate Candidate) RootPath() string {
	letter := strings.TrimSpace(strings.TrimSuffix(candidate.DriveLetter, ":"))
	if letter == "" {
		return ""
	}
	return strings.ToUpper(letter) + ":\\"
}

func (candidate Candidate) PreparationIssues(minimumFreeBytes int64) []string {
	var issues []string
	if candidate.RootPath() == "" {
		return []string{"o Windows não atribuiu uma letra a este pendrive"}
	}
	if candidate.IsSystem {
		issues = append(issues, "este pendrive contém o Windows que está em uso")
	}
	if candidate.FileSystem == "" {
		issues = append(issues, "não foi possível identificar o sistema de arquivos")
	} else if !strings.EqualFold(candidate.FileSystem, "NTFS") {
		issues = append(issues, fmt.Sprintf(
			"o sistema de arquivos é %s; esta versão precisa de NTFS",
			candidate.FileSystem,
		))
	}
	if minimumFreeBytes > 0 && candidate.FreeBytes < uint64(minimumFreeBytes) {
		issues = append(issues, fmt.Sprintf(
			"há %.1f GiB livres; são necessários pelo menos %.1f GiB",
			bytesInGiB(candidate.FreeBytes),
			bytesInGiB(uint64(minimumFreeBytes)),
		))
	}
	return issues
}

func bytesInGiB(size uint64) float64 {
	return float64(size) / (1 << 30)
}
