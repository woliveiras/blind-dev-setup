package bundle

import (
	"reflect"
	"testing"
)

func TestWindowsManifestIsValid(t *testing.T) {
	got, err := WindowsManifest()
	if err != nil {
		t.Fatalf("WindowsManifest() error = %v", err)
	}
	if got.Platform != "windows-x64" || len(got.Components) != 6 {
		t.Fatalf("WindowsManifest() = %#v", got)
	}
}

func TestWindowsManifestFlattensMiseBinaryDirectory(t *testing.T) {
	got, err := WindowsManifest()
	if err != nil {
		t.Fatalf("WindowsManifest() error = %v", err)
	}
	for _, component := range got.Components {
		if component.ID != "mise" {
			continue
		}
		if component.StripComponents != 2 {
			t.Errorf("mise StripComponents = %d, want 2", component.StripComponents)
		}
		if want := []string{"mise.exe"}; !reflect.DeepEqual(component.ExpectedFiles, want) {
			t.Errorf("mise ExpectedFiles = %#v, want %#v", component.ExpectedFiles, want)
		}
		return
	}
	t.Fatal("mise component is missing")
}
