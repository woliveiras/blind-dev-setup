# blind-dev-setup

Gerador de um ambiente de desenvolvimento portátil para pessoas cegas, com mensagens lineares e operação por teclado e leitor de tela.

A v0.1.0 prepara um pendrive NTFS para Windows 11 x64 com:

- Git for Windows Portable;
- Visual Studio Code Portable como editor principal;
- NVDA Portable;
- Notepad++ Portable como editor auxiliar;
- DBeaver Community Portable, ainda experimental quanto à acessibilidade;
- Python, Node.js, uv e pnpm em versões fixadas e instaladas pelo mise;
- configurações, caches, workspaces, exemplos e dados mantidos no pendrive;
- guias de NVDA, uv, pnpm, limitações e validação prática.

O código e o executável Windows foram validados automaticamente. O protocolo completo ainda precisa ser executado em Windows limpo por pessoas cegas; automação não prova acessibilidade.

## Segurança do destino

O gerador:

- exige um destino explícito;
- pode listar pendrives conectados, mas nunca formata, particiona ou escolhe uma unidade;
- recusa o disco do sistema;
- exige NTFS e 12 GiB livres;
- recusa sobrescrever uma pasta blind-dev-setup existente;
- monta tudo em uma pasta temporária na mesma unidade;
- publica a pasta final somente depois de concluir todas as etapas;
- baixa cada aplicativo da fonte oficial e confere tamanho e SHA-256;
- exige aceite explícito das licenças de terceiros.

Use um pendrive de pelo menos 32 GB. A preparação inicial baixa aproximadamente 607 MiB de aplicativos, além dos toolchains instalados pelo mise.

## Construção

Requisitos para compilar:

- Go 1.25 ou posterior;
- Git;
- Windows 11 x64 para executar prepare.

No PowerShell:

    go test ./...
    go build -trimpath -o blind-dev-setup-windows-x64.exe ./cmd/blind-dev-setup

Em macOS ou Linux é possível testar e gerar o executável Windows:

    go test ./...
    GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -o blind-dev-setup-windows-x64.exe ./cmd/blind-dev-setup

O script scripts/build.ps1 executa os testes, cria dist/blind-dev-setup-windows-x64.exe e mostra seu SHA-256.

## Uso

Coloque o executável em uma pasta e conecte o pendrive. Para abrir o PowerShell nessa pasta pelo teclado:

1. abra a pasta no Explorador de Arquivos;
2. pressione Ctrl+L;
3. digite powershell e pressione Enter.

Procure os pendrives conectados:

    .\blind-dev-setup-windows-x64.exe list-targets

O comando mostra nome, letra, nome do volume, sistema de arquivos, tamanho, espaço livre e situação de cada unidade USB. Ele não modifica nada e explica o próximo passo. Compare o nome e o tamanho apresentados com o pendrive conectado antes de usar a letra indicada.

Os exemplos seguintes usam E:. Troque E: pela letra mostrada para o seu pendrive.

Mostre o plano sem escrever:

    .\blind-dev-setup-windows-x64.exe plan --target E:\

Prepare um ambiente novo depois de ler as licenças exibidas:

    .\blind-dev-setup-windows-x64.exe prepare --target E:\ --accept-licenses

Um cache específico pode ser informado:

    .\blind-dev-setup-windows-x64.exe prepare --target E:\ --cache C:\blind-dev-cache --accept-licenses

Verifique uma preparação sem modificá-la:

    .\blind-dev-setup-windows-x64.exe verify --target E:\

Depois, abra E:\blind-dev-setup\START.cmd para iniciar NVDA e VS Code, ou DEV-SHELL.cmd para abrir o shell portátil.

Se `list-targets` não encontrar nada, desconecte o pendrive, conecte-o novamente, aguarde alguns segundos e repita o comando. Se o Windows não atribuir uma letra ou o sistema de arquivos não for NTFS, a ferramenta explicará o problema sem tentar formatar a unidade.

## Estrutura gerada

- tools contém os aplicativos portáteis;
- data contém mise, toolchains, store pnpm e workspace DBeaver;
- cache contém caches do mise e uv;
- config contém manifesto, mise, npm e ambiente dos launchers;
- workspace contém exemplos e a pasta projects;
- docs contém uso, limitações e protocolo de validação;
- READY.txt só existe em uma preparação concluída.

## Gestão dos ambientes

mise instala e fixa Python, Node.js, uv e pnpm. Dentro de um projeto Python, use uv sync e uv run. Em Node.js, use pnpm install e workspaces definidos por pnpm-workspace.yaml.

.venv e node_modules são derivados. Uma mudança da letra do pendrive pode exigir sua recriação a partir de pyproject.toml, uv.lock, package.json e pnpm-lock.yaml.

## Limites

NVDA Portable não cobre login, UAC, telas seguras ou aplicativos elevados. Para esses fluxos, use uma instalação permanente oficial do NVDA. O gerador não copia perfil pessoal, add-ons, credenciais, tokens ou chaves SSH.

VS Code não recebe extensões automaticamente na v0.1.0; elas exigem uma decisão separada de proveniência e validação. DBeaver precisa de validação empírica com NVDA. Atualização transacional e Linux estão no roadmap.

Leia [a decisão arquitetural](docs/adr/0001-windows-portable-builder.md), [os limites das evidências](docs/research/evidence-boundaries.md), o [roadmap](ROADMAP.md), as [diretrizes de contribuição](CONTRIBUTING.md) e a [política de segurança](SECURITY.md).

## Licenças

O código deste projeto usa licença MIT. As ferramentas instaladas mantêm suas próprias licenças. Seus identificadores, links, versões, URLs e hashes estão em internal/bundle/windows-x64.json; os binários de terceiros não são armazenados no Git.
