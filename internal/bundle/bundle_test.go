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
	if got.Platform != "windows-x64" || len(got.Components) != 7 {
		t.Fatalf("WindowsManifest() = %#v", got)
	}
}

func TestWindowsManifestInstallsConsoleExtractorBeforePortableGit(t *testing.T) {
	current, err := WindowsManifest()
	if err != nil {
		t.Fatal(err)
	}
	extractorReady := false
	for _, component := range current.Components {
		if component.ID == "sevenzip" {
			extractorReady = component.Kind == "executable" && component.Destination == "tools/7zip" && reflect.DeepEqual(component.ExpectedFiles, []string{"7zr.exe"})
		}
		if component.ID == "git" {
			if !extractorReady || component.Kind != "git-portable-7zip" {
				t.Fatal("Portable Git must use the verified console extractor installed earlier in the bundle")
			}
			return
		}
	}
	t.Fatal("Git component is missing")
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
