package prepare

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
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
	commands     []recordedCommand
	environments [][]string
	failAt       int
	err          error
}

func (r *recordingRunner) Run(_ context.Context, executable string, arguments []string, directory string, environment []string, _ io.Writer) error {
	r.environments = append(r.environments, environment)
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
	batch := writePostInstallBatch(t, root)
	t.Setenv("BASH_ENV", "must-not-source-user-code")
	t.Setenv("MSYSTEM", "UCRT64")
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
			executable: filepath.Join(destination, "usr", "bin", "bash.exe"),
			arguments:  []string{"--noprofile", "--norc", "-c", gitPostInstallScript, "blind-dev-git-post-install", "/etc/post-install"},
			directory:  destination,
		},
	}
	if !reflect.DeepEqual(runner.commands, want) {
		t.Fatalf("commands = %#v, want %#v", runner.commands, want)
	}
	if _, err := os.Stat(batch); !os.IsNotExist(err) {
		t.Fatalf("successful post-install retained batch: %v", err)
	}
	var msystemFound bool
	for _, entry := range runner.environments[1] {
		if strings.HasPrefix(strings.ToUpper(entry), "BASH_ENV=") || entry == "MSYSTEM=UCRT64" {
			t.Fatalf("unsafe inherited Bash environment: %s", entry)
		}
		msystemFound = msystemFound || entry == "MSYSTEM=MINGW64"
	}
	if !msystemFound {
		t.Fatal("missing pinned Git MSYSTEM")
	}
}

func TestPortableGitStopsOnExtractionOrPostInstallFailure(t *testing.T) {
	for _, failAt := range []int{1, 2} {
		runner := &recordingRunner{failAt: failAt, err: errors.New("child failed")}
		root := t.TempDir()
		batch := writePostInstallBatch(t, root)
		err := (Installer{Runner: runner}).Install(context.Background(), manifest.Component{
			ID: "git", Kind: "git-portable-7zip", Destination: "tools/git",
		}, "git.artifact", root, io.Discard)
		if _, statErr := os.Stat(batch); statErr != nil {
			t.Fatalf("failed post-install removed batch: %v", statErr)
		}
		if failAt == 2 && (err == nil || !strings.Contains(err.Error(), "bash.exe") || !strings.Contains(err.Error(), root)) {
			t.Fatalf("post-install error lacks executable/directory diagnostics: %v", err)
		}
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

func writePostInstallBatch(t *testing.T, root string) string {
	t.Helper()
	path := filepath.Join(root, "tools", "git", "post-install.bat")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("upstream post-install"), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestGitPostInstallSourcesHooksInOrderWithSharedStateAndDiagnostics(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("real bundled Bash is covered by the Windows NTFS installation test")
	}
	hooks := filepath.Join(t.TempDir(), "hooks with spaces")
	if err := os.Mkdir(hooks, 0o755); err != nil {
		t.Fatal(err)
	}
	for name, script := range map[string]string{
		"01-first.post":   "shared_value=kept; false\n",
		"02-second.post":  "test \"$shared_value\" = kept || exit 17; printf 'shared-state-ok\\n'\n",
		"99-cleanup.post": "rm -rf \"$1\"\n",
	} {
		if err := os.WriteFile(filepath.Join(hooks, name), []byte(script), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	command := exec.Command("bash", "--noprofile", "--norc", "-c", gitPostInstallScript, "test", hooks)
	command.Env = gitPostInstallEnvironment()
	output, err := command.CombinedOutput()
	if err != nil || !bytes.Contains(output, []byte("shared-state-ok")) || !bytes.Contains(output, []byte("status: 1")) {
		t.Fatalf("hook execution: %v, output: %s", err, output)
	}
	if _, err := os.Stat(hooks); !os.IsNotExist(err) {
		t.Fatalf("cleanup hook did not run: %v", err)
	}
}

func TestGitPostInstallPropagatesFatalHookFailure(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("real bundled Bash is covered by the Windows NTFS installation test")
	}
	hooks := t.TempDir()
	if err := os.WriteFile(filepath.Join(hooks, "01-failure.post"), []byte("echo fatal-hook >&2; exit 23\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	command := exec.Command("bash", "--noprofile", "--norc", "-c", gitPostInstallScript, "test", hooks)
	command.Env = gitPostInstallEnvironment()
	output, err := command.CombinedOutput()
	var exitError *exec.ExitError
	if !errors.As(err, &exitError) || exitError.ExitCode() != 23 || !bytes.Contains(output, []byte("01-failure.post")) || !bytes.Contains(output, []byte("fatal-hook")) {
		t.Fatalf("fatal hook: %v, output: %s", err, output)
	}
}
