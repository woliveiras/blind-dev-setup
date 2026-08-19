package bundle

import "testing"

func TestWindowsManifestIsValid(t *testing.T) {
	got, err := WindowsManifest()
	if err != nil {
		t.Fatalf("WindowsManifest() error = %v", err)
	}
	if got.Platform != "windows-x64" || len(got.Components) != 6 {
		t.Fatalf("WindowsManifest() = %#v", got)
	}
}
