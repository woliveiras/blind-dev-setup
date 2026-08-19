package target

import (
	"strings"
	"testing"
)

func TestValidateAcceptsNewNTFSTargetWithEnoughSpace(t *testing.T) {
	details := Details{
		Path:              "E:\\",
		Volume:            "E:",
		SystemVolume:      "C:",
		FileSystem:        "NTFS",
		FreeBytes:         20 << 30,
		DestinationExists: false,
	}
	if err := Validate(details, 12<<30); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestValidateRejectsUnsafeTargets(t *testing.T) {
	tests := []struct {
		name    string
		details Details
		want    string
	}{
		{
			name:    "system volume",
			details: Details{Path: "C:\\", Volume: "C:", SystemVolume: "C:", FileSystem: "NTFS", FreeBytes: 20 << 30},
			want:    "sistema",
		},
		{
			name:    "non NTFS",
			details: Details{Path: "E:\\", Volume: "E:", SystemVolume: "C:", FileSystem: "exFAT", FreeBytes: 20 << 30},
			want:    "NTFS",
		},
		{
			name:    "existing installation",
			details: Details{Path: "E:\\", Volume: "E:", SystemVolume: "C:", FileSystem: "NTFS", FreeBytes: 20 << 30, DestinationExists: true},
			want:    "já existe",
		},
		{
			name:    "insufficient space",
			details: Details{Path: "E:\\", Volume: "E:", SystemVolume: "C:", FileSystem: "NTFS", FreeBytes: 2 << 30},
			want:    "espaço",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.details, 12<<30)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("Validate() error = %v, want substring %q", err, tt.want)
			}
		})
	}
}
