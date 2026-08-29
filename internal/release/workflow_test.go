package release_test

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
)

func TestReleasePleaseTracksOneGoPackageAndEmbeddedVersion(t *testing.T) {
	root := projectRoot(t)
	var config struct {
		BootstrapSHA string `json:"bootstrap-sha"`
		Packages     map[string]struct {
			ReleaseType           string `json:"release-type"`
			PackageName           string `json:"package-name"`
			IncludeComponentInTag bool   `json:"include-component-in-tag"`
			IncludeVInTag         bool   `json:"include-v-in-tag"`
			Draft                 bool   `json:"draft"`
			ExtraFiles            []struct {
				Type     string `json:"type"`
				Path     string `json:"path"`
				JSONPath string `json:"jsonpath"`
			} `json:"extra-files"`
		} `json:"packages"`
	}
	if err := json.Unmarshal(
		[]byte(readFile(t, filepath.Join(root, "release-please-config.json"))),
		&config,
	); err != nil {
		t.Fatalf("decode release-please-config.json: %v", err)
	}
	if len(config.BootstrapSHA) != 40 {
		t.Errorf("bootstrap-sha = %q", config.BootstrapSHA)
	}
	if len(config.Packages) != 1 {
		t.Fatalf("packages = %#v", config.Packages)
	}
	pkg, ok := config.Packages["."]
	if !ok {
		t.Fatal("root package is not configured")
	}
	if pkg.ReleaseType != "go" || pkg.PackageName != "blind-dev-setup" {
		t.Errorf("root package = %#v", pkg)
	}
	if pkg.IncludeComponentInTag || !pkg.IncludeVInTag {
		t.Errorf("tag configuration = %#v", pkg)
	}
	if !pkg.Draft {
		t.Error("GitHub Release is public before its artifacts are verified and attached")
	}
	if len(pkg.ExtraFiles) != 1 ||
		pkg.ExtraFiles[0].Type != "json" ||
		pkg.ExtraFiles[0].Path != "internal/bundle/windows-x64.json" ||
		pkg.ExtraFiles[0].JSONPath != "$.release" {
		t.Errorf("extra-files = %#v", pkg.ExtraFiles)
	}

	var manifest map[string]string
	if err := json.Unmarshal(
		[]byte(readFile(t, filepath.Join(root, ".release-please-manifest.json"))),
		&manifest,
	); err != nil {
		t.Fatalf("decode .release-please-manifest.json: %v", err)
	}
	if len(manifest) != 1 {
		t.Fatalf("release manifest = %#v", manifest)
	}
	releaseVersion, ok := manifest["."]
	if !ok || releaseVersion == "" {
		t.Fatalf("root package release version = %q", releaseVersion)
	}

	// Release Please updates both files for every release; their versions must
	// stay synchronized without pinning this test to a specific release number.
	var bundle struct {
		Release string `json:"release"`
	}
	if err := json.Unmarshal(
		[]byte(readFile(t, filepath.Join(root, "internal", "bundle", "windows-x64.json"))),
		&bundle,
	); err != nil {
		t.Fatalf("decode internal/bundle/windows-x64.json: %v", err)
	}
	if bundle.Release != releaseVersion {
		t.Errorf("embedded release version = %q, want %q", bundle.Release, releaseVersion)
	}
}

func TestReleaseWorkflowPublishesOnlyTestedWindowsArtifacts(t *testing.T) {
	root := projectRoot(t)
	workflow := readFile(t, filepath.Join(root, ".github", "workflows", "release-please.yml"))
	for _, required := range []string{
		"push:",
		"workflow_dispatch:",
		"permissions: {}",
		"contents: write",
		"issues: write",
		"pull-requests: write",
		"googleapis/release-please-action@45996ed1f6d02564a971a2fa1b5860e934307cf7",
		"release_created:",
		"tag_name:",
		"release_sha:",
		"release_version:",
		"needs.release-please.outputs.release_created == 'true'",
		"ref: ${{ needs.release-please.outputs.release_sha }}",
		"persist-credentials: false",
		"./scripts/build.ps1",
		"blind-dev-setup-windows-x64.exe",
		"blind-dev-setup-windows-x64.zip",
		"SHA256SUMS.txt",
		"gh release upload",
		"gh release edit",
		"--draft=false",
	} {
		if !strings.Contains(workflow, required) {
			t.Errorf("release workflow does not contain %q", required)
		}
	}
	if strings.Contains(workflow, "pull_request_target") {
		t.Error("release workflow executes with write authority on pull_request_target")
	}
	if strings.Count(workflow, "GH_TOKEN:") != 1 {
		t.Error("release token must only be exposed to the asset publication step")
	}
}
