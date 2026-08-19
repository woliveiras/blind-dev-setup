package manifest

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestParseValidManifest(t *testing.T) {
	want := validManifest()
	raw, err := json.Marshal(want)
	if err != nil {
		t.Fatal(err)
	}

	got, err := Parse(bytes.NewReader(raw))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if got.Release != want.Release || got.Components[0].ID != want.Components[0].ID {
		t.Fatalf("Parse() = %#v", got)
	}
}

func TestParseRejectsUnsafeOrUnverifiableComponents(t *testing.T) {
	tests := []struct {
		name string
		edit func(*Manifest)
		want string
	}{
		{name: "plain HTTP", edit: func(m *Manifest) { m.Components[0].URL = "http://example.test/tool.zip" }, want: "HTTPS"},
		{name: "invalid checksum", edit: func(m *Manifest) { m.Components[0].SHA256 = "unknown" }, want: "SHA-256"},
		{name: "absolute destination", edit: func(m *Manifest) { m.Components[0].Destination = "C:\\tool" }, want: "relative"},
		{name: "parent destination", edit: func(m *Manifest) { m.Components[0].Destination = "../tool" }, want: "relative"},
		{name: "unknown installer", edit: func(m *Manifest) { m.Components[0].Kind = "powershell" }, want: "kind"},
		{name: "duplicate component", edit: func(m *Manifest) { m.Components = append(m.Components, m.Components[0]) }, want: "duplicate"},
		{
			name: "overlapping destinations",
			edit: func(m *Manifest) {
				second := m.Components[0]
				second.ID = "nested"
				second.Destination = "tools/editor/extensions"
				m.Components = append(m.Components, second)
			},
			want: "overlap",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			candidate := validManifest()
			tt.edit(&candidate)
			if err := candidate.Validate(); err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("Validate() error = %v, want substring %q", err, tt.want)
			}
		})
	}
}

func validManifest() Manifest {
	return Manifest{
		SchemaVersion:    1,
		Release:          "0.1.0",
		Platform:         "windows-x64",
		MinimumFreeBytes: 1024,
		Toolchain: map[string]string{
			"python": "3.14.7",
			"node":   "24.19.0",
			"uv":     "0.12.5",
			"pnpm":   "11.22.0",
		},
		Components: []Component{{
			ID:            "editor",
			Name:          "Editor",
			Version:       "1.2.3",
			Required:      true,
			URL:           "https://example.test/editor.zip",
			SHA256:        "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			License:       "Example",
			LicenseURL:    "https://example.test/license",
			Kind:          "zip",
			Destination:   "tools/editor",
			DownloadBytes: 100,
			ExpectedFiles: []string{"editor.exe"},
		}},
	}
}
