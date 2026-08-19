package main

import (
	"fmt"
	"os"

	"github.com/woliveiras/blind-dev-setup/internal/bundle"
	"github.com/woliveiras/blind-dev-setup/internal/cli"
)

func main() {
	current, err := bundle.WindowsManifest()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro interno: manifesto inválido: %v\n", err)
		os.Exit(1)
	}
	os.Exit(cli.Run(os.Args[1:], os.Stdout, os.Stderr, cli.Dependencies{Manifest: current}))
}
