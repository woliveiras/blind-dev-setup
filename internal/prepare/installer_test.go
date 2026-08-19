package prepare

import (
	"context"
	"io"
	"path/filepath"
	"reflect"
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
}

func (r *recordingRunner) Run(_ context.Context, executable string, arguments []string, directory string, _ []string, _ io.Writer) error {
	r.commands = append(r.commands, recordedCommand{executable: executable, arguments: arguments, directory: directory})
	return nil
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
