package cli

import (
	"flag"
	"fmt"
	"io"
	"strings"

	"github.com/woliveiras/blind-dev-setup/internal/manifest"
)

type PrepareRequest struct {
	Target string
	Cache  string
	Output io.Writer
}

type Dependencies struct {
	Manifest manifest.Manifest
	Prepare  func(PrepareRequest) error
	Verify   func(target string, output io.Writer) error
}

func Run(args []string, stdout, stderr io.Writer, dependencies Dependencies) int {
	if len(args) == 0 {
		printUsage(stderr)
		return 2
	}

	switch args[0] {
	case "help", "--help", "-h":
		printUsage(stdout)
		return 0
	case "version":
		fmt.Fprintf(stdout, "blind-dev-setup %s\n", dependencies.Manifest.Release)
		return 0
	case "plan":
		return runPlan(args[1:], stdout, stderr, dependencies.Manifest)
	case "prepare":
		return runPrepare(args[1:], stdout, stderr, dependencies)
	case "verify":
		return runVerify(args[1:], stdout, stderr, dependencies)
	default:
		fmt.Fprintf(stderr, "Erro: Comando desconhecido: %s\n\n", args[0])
		printUsage(stderr)
		return 2
	}
}

func runPlan(args []string, stdout, stderr io.Writer, current manifest.Manifest) int {
	flags := flag.NewFlagSet("plan", flag.ContinueOnError)
	flags.SetOutput(stderr)
	target := flags.String("target", "", "unidade ou diretório de destino")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if strings.TrimSpace(*target) == "" {
		fmt.Fprintln(stderr, "Erro: --target é obrigatório.")
		return 2
	}

	fmt.Fprintf(stdout, "blind-dev-setup %s\n", current.Release)
	fmt.Fprintln(stdout, "Plano de preparação para Windows 11 x64")
	fmt.Fprintf(stdout, "Destino: %s\n", *target)
	fmt.Fprintf(stdout, "Espaço livre mínimo: %.1f GiB\n", float64(current.MinimumFreeBytes)/(1024*1024*1024))
	fmt.Fprintln(stdout, "Componentes:")
	var total int64
	for _, component := range current.Components {
		fmt.Fprintf(stdout, "- %s %s; licença: %s\n", component.Name, component.Version, component.License)
		total += component.DownloadBytes
	}
	fmt.Fprintf(stdout, "Download estimado dos aplicativos: %.1f MiB\n", float64(total)/(1024*1024))
	fmt.Fprintln(stdout, "Toolchain instalado pelo mise:")
	for _, tool := range []string{"python", "node", "uv", "pnpm"} {
		fmt.Fprintf(stdout, "- %s %s\n", toolName(tool), current.Toolchain[tool])
	}
	fmt.Fprintln(stdout, "Este comando é somente leitura; nenhum arquivo será modificado.")
	return 0
}

func runPrepare(args []string, stdout, stderr io.Writer, dependencies Dependencies) int {
	flags := flag.NewFlagSet("prepare", flag.ContinueOnError)
	flags.SetOutput(stderr)
	target := flags.String("target", "", "unidade ou diretório de destino")
	cache := flags.String("cache", "", "diretório opcional para downloads verificados")
	accept := flags.Bool("accept-licenses", false, "aceita as licenças das ferramentas de terceiros")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if strings.TrimSpace(*target) == "" {
		fmt.Fprintln(stderr, "Erro: --target é obrigatório.")
		return 2
	}
	if !*accept {
		fmt.Fprintln(stderr, "Erro: leia as licenças mostradas por plan e informe --accept-licenses para continuar.")
		return 2
	}
	if dependencies.Prepare == nil {
		fmt.Fprintln(stderr, "Erro: preparação indisponível nesta compilação.")
		return 1
	}
	if err := dependencies.Prepare(PrepareRequest{Target: *target, Cache: *cache, Output: stdout}); err != nil {
		fmt.Fprintf(stderr, "Erro: %v\n", err)
		return 1
	}
	return 0
}

func runVerify(args []string, stdout, stderr io.Writer, dependencies Dependencies) int {
	flags := flag.NewFlagSet("verify", flag.ContinueOnError)
	flags.SetOutput(stderr)
	target := flags.String("target", "", "unidade ou diretório de destino")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if strings.TrimSpace(*target) == "" {
		fmt.Fprintln(stderr, "Erro: --target é obrigatório.")
		return 2
	}
	if dependencies.Verify == nil {
		fmt.Fprintln(stderr, "Erro: verificação indisponível nesta compilação.")
		return 1
	}
	if err := dependencies.Verify(*target, stdout); err != nil {
		fmt.Fprintf(stderr, "Erro: %v\n", err)
		return 1
	}
	return 0
}

func toolName(id string) string {
	names := map[string]string{"python": "Python", "node": "Node.js", "uv": "uv", "pnpm": "pnpm"}
	if name, ok := names[id]; ok {
		return name
	}
	return id
}

func printUsage(output io.Writer) {
	lines := []string{
		"Uso: blind-dev-setup <comando> [opções]",
		"",
		"Comandos:",
		"  plan     mostra componentes, licenças e ações sem escrever",
		"  prepare  prepara um ambiente novo no destino explícito",
		"  verify   verifica uma preparação existente sem modificá-la",
		"  version  mostra a versão do gerador",
		"  help     mostra esta ajuda",
	}
	fmt.Fprintln(output, strings.Join(lines, "\n"))
}
