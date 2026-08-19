package manifest

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"path"
	"regexp"
	"strings"
)

const CurrentSchemaVersion = 1

var (
	componentIDPattern = regexp.MustCompile("^[a-z][a-z0-9-]*$")
	sha256Pattern      = regexp.MustCompile("^[a-f0-9]{64}$")
	versionPattern     = regexp.MustCompile("^[A-Za-z0-9][A-Za-z0-9.+_-]*$")
	windowsVolume      = regexp.MustCompile("^[A-Za-z]:")
)

type Manifest struct {
	SchemaVersion    int               `json:"schema_version"`
	Release          string            `json:"release"`
	Platform         string            `json:"platform"`
	MinimumFreeBytes int64             `json:"minimum_free_bytes"`
	Toolchain        map[string]string `json:"toolchain"`
	Components       []Component       `json:"components"`
}

type Component struct {
	ID              string   `json:"id"`
	Name            string   `json:"name"`
	Version         string   `json:"version"`
	Required        bool     `json:"required"`
	URL             string   `json:"url"`
	SHA256          string   `json:"sha256"`
	License         string   `json:"license"`
	LicenseURL      string   `json:"license_url"`
	Kind            string   `json:"kind"`
	Destination     string   `json:"destination"`
	DownloadBytes   int64    `json:"download_bytes"`
	StripComponents int      `json:"strip_components,omitempty"`
	Arguments       []string `json:"arguments,omitempty"`
	ExpectedFiles   []string `json:"expected_files"`
}

func Parse(reader io.Reader) (Manifest, error) {
	decoder := json.NewDecoder(reader)
	decoder.DisallowUnknownFields()

	var result Manifest
	if err := decoder.Decode(&result); err != nil {
		return Manifest{}, fmt.Errorf("decode manifest: %w", err)
	}
	if err := ensureEnd(decoder); err != nil {
		return Manifest{}, err
	}
	if err := result.Validate(); err != nil {
		return Manifest{}, err
	}
	return result, nil
}

func ensureEnd(decoder *json.Decoder) error {
	var extra any
	if err := decoder.Decode(&extra); err == io.EOF {
		return nil
	} else if err != nil {
		return fmt.Errorf("decode manifest trailer: %w", err)
	}
	return errors.New("manifest contains more than one JSON value")
}

func (m Manifest) Validate() error {
	if m.SchemaVersion != CurrentSchemaVersion {
		return fmt.Errorf("unsupported schema_version %d", m.SchemaVersion)
	}
	if strings.TrimSpace(m.Release) == "" {
		return errors.New("release is required")
	}
	if m.Platform != "windows-x64" {
		return fmt.Errorf("unsupported platform %q", m.Platform)
	}
	if m.MinimumFreeBytes <= 0 {
		return errors.New("minimum_free_bytes must be positive")
	}
	for _, tool := range []string{"python", "node", "uv", "pnpm"} {
		if !versionPattern.MatchString(m.Toolchain[tool]) {
			return fmt.Errorf("toolchain version for %s is required", tool)
		}
	}
	if len(m.Components) == 0 {
		return errors.New("at least one component is required")
	}

	ids := make(map[string]struct{}, len(m.Components))
	for index, component := range m.Components {
		if err := component.validate(); err != nil {
			return fmt.Errorf("component %d: %w", index+1, err)
		}
		if _, exists := ids[component.ID]; exists {
			return fmt.Errorf("duplicate component id %q", component.ID)
		}
		ids[component.ID] = struct{}{}
	}
	return nil
}

func (c Component) validate() error {
	if !componentIDPattern.MatchString(c.ID) {
		return fmt.Errorf("invalid id %q", c.ID)
	}
	if strings.TrimSpace(c.Name) == "" || !versionPattern.MatchString(c.Version) {
		return errors.New("name and version are required")
	}
	if err := validateHTTPS(c.URL, "download URL"); err != nil {
		return err
	}
	if !sha256Pattern.MatchString(c.SHA256) {
		return errors.New("SHA-256 must contain 64 lowercase hexadecimal characters")
	}
	if strings.TrimSpace(c.License) == "" {
		return errors.New("license is required")
	}
	if err := validateHTTPS(c.LicenseURL, "license URL"); err != nil {
		return err
	}
	switch c.Kind {
	case "zip", "self-extracting-7zip", "nvda-portable", "executable":
	default:
		return fmt.Errorf("unsupported installation kind %q", c.Kind)
	}
	if !safeRelativePath(c.Destination) {
		return fmt.Errorf("destination %q must be a safe relative path", c.Destination)
	}
	if c.DownloadBytes <= 0 {
		return errors.New("download_bytes must be positive")
	}
	if c.StripComponents < 0 {
		return errors.New("strip_components cannot be negative")
	}
	if len(c.ExpectedFiles) == 0 {
		return errors.New("expected_files cannot be empty")
	}
	for _, expected := range c.ExpectedFiles {
		if !safeRelativePath(expected) {
			return fmt.Errorf("expected file %q must be a safe relative path", expected)
		}
	}
	return nil
}

func validateHTTPS(raw, label string) error {
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Scheme != "https" || parsed.Host == "" || parsed.User != nil {
		return fmt.Errorf("%s must use HTTPS without embedded credentials", label)
	}
	return nil
}

func safeRelativePath(value string) bool {
	if value == "" || windowsVolume.MatchString(value) {
		return false
	}
	normalized := strings.ReplaceAll(value, "\\", "/")
	cleaned := path.Clean(normalized)
	return cleaned != "." && cleaned != ".." && !strings.HasPrefix(cleaned, "../") && !strings.HasPrefix(cleaned, "/")
}
