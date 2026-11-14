package prepare

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/woliveiras/blind-dev-setup/internal/manifest"
)

type recordedCommand struct {
	executable string
	arguments  []string
	directory  string
}

type recordingRunner struct {
	commands []recordedCommand
	failAt   int
	err      error
}

func (r *recordingRunner) Run(_ context.Context, executable string, arguments []string, directory string, _ []string, _ io.Writer) error {
	r.commands = append(r.commands, recordedCommand{executable: executable, arguments: arguments, directory: directory})
	if r.err != nil && len(r.commands) == r.failAt {
		return r.err
	}
	return nil
}

func TestPortableGitUsesBundledConsoleExtractorAndPreservesUpstreamPostInstall(t *testing.T) {
	runner := &recordingRunner{}
	installer := Installer{Runner: runner}
	root := filepath.Join(t.TempDir(), "drive with spaces & symbols")
	component := manifest.Component{ID: "git", Kind: "git-portable-7zip", Destination: "tools/git"}
	// A relative cache is resolved before the extractor changes its directory.
	artifact := filepath.Join("relative cache", "verified-git.artifact")
	if err := installer.Install(context.Background(), component, artifact, root, io.Discard); err != nil {
		t.Fatal(err)
	}
	absoluteArtifact, err := filepath.Abs(artifact)
	if err != nil {
		t.Fatal(err)
	}
	destination := filepath.Join(root, "tools", "git")
	want := []recordedCommand{
		{
			executable: filepath.Join(root, "tools", "7zip", "7zr.exe"),
			arguments:  []string{"x", "-t7z", "-y", "-bb0", "-bsp0", "-o" + destination, absoluteArtifact},
			directory:  destination,
		},
		{
			executable: filepath.Join(destination, "git-bash.exe"),
			arguments:  []string{"--needs-console", "--hide", "--no-cd", "--command=post-install.bat"},
			directory:  destination,
		},
	}
	if !reflect.DeepEqual(runner.commands, want) {
		t.Fatalf("commands = %#v, want %#v", runner.commands, want)
	}
}

func TestPortableGitStopsOnExtractionOrPostInstallFailure(t *testing.T) {
	for _, failAt := range []int{1, 2} {
		runner := &recordingRunner{failAt: failAt, err: errors.New("child failed")}
		err := (Installer{Runner: runner}).Install(context.Background(), manifest.Component{
			ID: "git", Kind: "git-portable-7zip", Destination: "tools/git",
		}, "git.artifact", t.TempDir(), io.Discard)
		if err == nil || !strings.Contains(err.Error(), "child failed") || len(runner.commands) != failAt {
			t.Fatalf("failure %d: error=%v commands=%v", failAt, err, runner.commands)
		}
	}
}

func TestPortableGitRejectsOtherComponents(t *testing.T) {
	for _, id := range []string{"other", ""} {
		runner := &recordingRunner{}
		err := (Installer{Runner: runner}).Install(context.Background(), manifest.Component{
			ID: id, Kind: "git-portable-7zip", Destination: "tools/git",
		}, "git.artifact", t.TempDir(), io.Discard)
		if err == nil || len(runner.commands) != 0 {
			t.Fatalf("component %s: error=%v commands=%v", id, err, runner.commands)
		}
	}
}

func TestExecutableUsesManifestFilenameInsteadOfCacheFilename(t *testing.T) {
	root := t.TempDir()
	artifact := filepath.Join(t.TempDir(), "checksum.artifact")
	if err := os.WriteFile(artifact, []byte("verified executable"), 0o644); err != nil {
		t.Fatal(err)
	}
	component := manifest.Component{Kind: "executable", Destination: "tools/7zip", ExpectedFiles: []string{"7zr.exe"}}
	installer := Installer{}
	if err := installer.Install(context.Background(), component, artifact, root, io.Discard); err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(filepath.Join(root, "tools", "7zip", "7zr.exe"))
	if err != nil || string(content) != "verified executable" {
		t.Fatalf("installed executable = %q, error=%v", content, err)
	}
	if err := installer.Install(context.Background(), component, artifact, root, io.Discard); err == nil {
		t.Fatal("executable overwrite was accepted")
	}
}

func TestExecutableRejectsUnsafeOrAmbiguousFilenames(t *testing.T) {
	for _, files := range [][]string{nil, {"one.exe", "two.exe"}, {"../outside.exe"}, {"..\\outside.exe"}, {"nested/file.exe"}, {"file.exe:stream"}, {"."}, {""}} {
		err := (Installer{}).Install(context.Background(), manifest.Component{
			Kind: "executable", Destination: "tools/7zip", ExpectedFiles: files,
		}, "checksum.artifact", t.TempDir(), io.Discard)
		if err == nil || !strings.Contains(err.Error(), "nome de arquivo") {
			t.Fatalf("files=%v error=%v", files, err)
		}
	}
}

func TestInstallerUsesArgumentListForSelfExtractingArchive(t *testing.T) {
	runner := &recordingRunner{}
	installer := Installer{Runner: runner}
	root := t.TempDir()
	component := manifest.Component{
		Kind:        "self-extracting-7zip",
		Destination: "tools/git",
		Arguments:   []string{"-y", "-o{destination}"},
	}

	if err := installer.Install(context.Background(), component, "git.exe", root, io.Discard); err != nil {
		t.Fatalf("Install() error = %v", err)
	}
	wantDestination := filepath.Join(root, "tools", "git")
	wantArguments := []string{"-y", "-o" + wantDestination}
	if len(runner.commands) != 1 || !reflect.DeepEqual(runner.commands[0].arguments, wantArguments) {
		t.Fatalf("commands = %#v, want arguments %#v", runner.commands, wantArguments)
	}
}

func TestInstallerUsesOfficialNVDAPortableArguments(t *testing.T) {
	runner := &recordingRunner{}
	installer := Installer{Runner: runner}
	root := t.TempDir()
	component := manifest.Component{Kind: "nvda-portable", Destination: "tools/nvda"}

	if err := installer.Install(context.Background(), component, "nvda.exe", root, io.Discard); err != nil {
		t.Fatalf("Install() error = %v", err)
	}
	want := []string{"--create-portable-silent", "--portable-path=" + filepath.Join(root, "tools", "nvda")}
	if len(runner.commands) != 1 || !reflect.DeepEqual(runner.commands[0].arguments, want) {
		t.Fatalf("commands = %#v, want arguments %#v", runner.commands, want)
	}
}
