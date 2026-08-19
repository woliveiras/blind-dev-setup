package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"

	"github.com/woliveiras/blind-dev-setup/internal/artifact"
	"github.com/woliveiras/blind-dev-setup/internal/bundle"
	"github.com/woliveiras/blind-dev-setup/internal/cli"
	"github.com/woliveiras/blind-dev-setup/internal/prepare"
	"github.com/woliveiras/blind-dev-setup/internal/target"
	"github.com/woliveiras/blind-dev-setup/internal/templates"
	"github.com/woliveiras/blind-dev-setup/internal/verify"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	current, err := bundle.WindowsManifest()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro interno: manifesto inválido: %v\n", err)
		os.Exit(1)
	}
	runner := prepare.ExecRunner{}
	creator := prepare.Creator{
		Manifest:     current,
		ManifestJSON: bundle.WindowsManifestJSON(),
		Fetch: artifact.Fetcher{Client: &http.Client{
			CheckRedirect: secureRedirect,
		}},
		Install:     prepare.Installer{Runner: runner},
		Materialize: templates.Writer{Manifest: current},
		Toolchain:   prepare.MiseToolchain{Runner: runner},
		Inspect:     target.Inspect,
	}
	dependencies := cli.Dependencies{
		Manifest: current,
		Prepare: func(request cli.PrepareRequest) error {
			return creator.Run(ctx, request.Target, request.Cache, request.Output)
		},
		Verify: verify.Run,
	}
	os.Exit(cli.Run(os.Args[1:], os.Stdout, os.Stderr, dependencies))
}

func secureRedirect(request *http.Request, previous []*http.Request) error {
	if len(previous) >= 10 {
		return fmt.Errorf("muitos redirecionamentos durante o download")
	}
	if request.URL.Scheme != "https" {
		return fmt.Errorf("redirecionamento de download para protocolo não seguro")
	}
	return nil
}
