package cli

import (
	"flag"
	"fmt"
	"io"
	"strings"

	"github.com/woliveiras/blind-dev-setup/internal/manifest"
	"github.com/woliveiras/blind-dev-setup/internal/target"
)

type PrepareRequest struct {
	Target string
	Cache  string
	Output io.Writer
}

type Dependencies struct {
	Manifest    manifest.Manifest
	ListTargets func() ([]target.Candidate, error)
	Prepare     func(PrepareRequest) error
	Verify      func(target string, output io.Writer) error
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
	case "list-targets":
		return runListTargets(stdout, stderr, dependencies)
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

func runListTargets(stdout, stderr io.Writer, dependencies Dependencies) int {
	if dependencies.ListTargets == nil {
		fmt.Fprintln(stderr, "Erro: a procura de pendrives não está disponível nesta compilação.")
		return 1
	}

	fmt.Fprintln(stdout, "Procurando pendrives conectados por USB...")
	candidates, err := dependencies.ListTargets()
	if err != nil {
		fmt.Fprintln(stderr, "Erro: não foi possível procurar pendrives.")
		fmt.Fprintf(stderr, "Detalhes: %v\n", err)
		fmt.Fprintln(stderr, "Feche e abra o programa novamente. Se o problema continuar, reinicie o Windows.")
		return 1
	}
	if len(candidates) == 0 {
		fmt.Fprintln(stdout, "Nenhum pendrive foi encontrado.")
		fmt.Fprintln(stdout, "Conecte o pendrive, aguarde alguns segundos e execute este comando novamente:")
		fmt.Fprintln(stdout, `.\blind-dev-setup-windows-x64.exe list-targets`)
		return 0
	}

	fmt.Fprintln(stdout)
	if len(candidates) == 1 {
		fmt.Fprintln(stdout, "1 pendrive encontrado.")
	} else {
		fmt.Fprintf(stdout, "%d pendrives encontrados.\n", len(candidates))
	}

	for index, candidate := range candidates {
		printCandidate(stdout, index+1, candidate, dependencies.Manifest.MinimumFreeBytes)
	}
	return 0
}

func printCandidate(output io.Writer, number int, candidate target.Candidate, minimumFreeBytes int64) {
	fmt.Fprintf(output, "\nPendrive %d\n", number)
	fmt.Fprintf(output, "Nome: %s\n", valueOr(candidate.Model, "não informado pelo dispositivo"))
	root := candidate.RootPath()
	if root == "" {
		fmt.Fprintln(output, "Letra: não atribuída pelo Windows")
	} else {
		fmt.Fprintf(output, "Letra: %s\n", root)
	}
	if candidate.Label != "" {
		fmt.Fprintf(output, "Nome do volume: %s\n", candidate.Label)
	}
	if candidate.FileSystem != "" {
		fmt.Fprintf(output, "Sistema de arquivos: %s\n", candidate.FileSystem)
	}
	if candidate.SizeBytes > 0 {
		fmt.Fprintf(output, "Tamanho total: %.1f GiB\n", bytesInGiB(candidate.SizeBytes))
	}
	if root != "" {
		fmt.Fprintf(output, "Espaço livre: %.1f GiB\n", bytesInGiB(candidate.FreeBytes))
	}

	issues := candidate.PreparationIssues(minimumFreeBytes)
	if len(issues) > 0 {
		fmt.Fprintln(output, "Situação: precisa de atenção")
		for _, issue := range issues {
			fmt.Fprintf(output, "Motivo: %s.\n", issue)
		}
		fmt.Fprintln(output, "Nenhuma alteração foi feita neste pendrive.")
		printCandidateGuidance(output, candidate, minimumFreeBytes)
		return
	}
	if candidate.DestinationExists {
		fmt.Fprintln(output, "Situação: o ambiente já está preparado")
		fmt.Fprintln(output, "Para verificar se está tudo certo, execute:")
		fmt.Fprintf(output, `.\blind-dev-setup-windows-x64.exe verify --target %s`+"\n", root)
		return
	}

	fmt.Fprintln(output, "Situação: pronto para preparar")
	fmt.Fprintln(output, "Próximo passo: veja o que será instalado, sem alterar o pendrive:")
	fmt.Fprintf(output, `.\blind-dev-setup-windows-x64.exe plan --target %s`+"\n", root)
}

func printCandidateGuidance(output io.Writer, candidate target.Candidate, minimumFreeBytes int64) {
	if candidate.RootPath() == "" {
		fmt.Fprintln(output, "Tente desconectar e conectar o pendrive novamente. Depois, execute list-targets outra vez.")
		return
	}
	if candidate.IsSystem {
		fmt.Fprintln(output, "Use outro pendrive. O programa nunca prepara a unidade que contém o Windows em uso.")
		return
	}
	if candidate.FileSystem != "" && !strings.EqualFold(candidate.FileSystem, "NTFS") {
		fmt.Fprintln(output, "Este programa não formata pendrives.")
		fmt.Fprintln(output, "Use outro pendrive NTFS ou peça ajuda para preparar este. Formatar apaga os arquivos; faça uma cópia antes.")
		return
	}
	if minimumFreeBytes > 0 && candidate.FreeBytes < uint64(minimumFreeBytes) {
		fmt.Fprintln(output, "Libere espaço depois de salvar os arquivos importantes em outro lugar, ou use outro pendrive.")
	}
}

func valueOr(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func bytesInGiB(size uint64) float64 {
	return float64(size) / (1 << 30)
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
		"Primeiro uso:",
		`  .\blind-dev-setup-windows-x64.exe list-targets`,
		"",
		"Comandos:",
		"  list-targets  procura pendrives e mostra qual letra usar",
		"  plan     mostra componentes, licenças e ações sem escrever",
		"  prepare  prepara um ambiente novo no destino explícito",
		"  verify   verifica uma preparação existente sem modificá-la",
		"  version  mostra a versão do gerador",
		"  help     mostra esta ajuda",
	}
	fmt.Fprintln(output, strings.Join(lines, "\n"))
}
