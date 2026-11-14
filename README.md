# blind-dev-setup

Gerador de um ambiente de desenvolvimento portátil para pessoas cegas, com assistente por teclado para uso com ou sem leitor de tela.

A versão atual prepara um pendrive NTFS para Windows 11 x64 com:

- Git for Windows Portable, extraído com um helper oficial do 7-Zip verificado;
- Visual Studio Code Portable como editor principal;
- NVDA Portable;
- Notepad++ Portable como editor auxiliar;
- DBeaver Community Portable, ainda experimental quanto à acessibilidade;
- Python, Node.js, uv e pnpm em versões fixadas e instaladas pelo mise;
- configurações, caches, workspaces, exemplos e dados mantidos no pendrive;
- guias de NVDA, uv, pnpm, limitações e validação prática.

O projeto tem testes automatizados e uma CI que prepara um ambiente real em volume NTFS temporário no Windows. Confira o resultado da revisão que está usando: um build ou a presença de arquivos não comprovam o uso completo. O protocolo por teclado e NVDA ainda exige validação prática com pessoas cegas; automação não prova acessibilidade.

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

## Baixar e usar

O programa já é distribuído pronto para Windows 11 x64. Você não precisa instalar Go, Git nem outras ferramentas e não precisa compilar o projeto.

### Download pela página do GitHub

1. abra a [página da release mais recente](https://github.com/woliveiras/blind-dev-setup/releases/latest);
2. em Assets, baixe `blind-dev-setup-windows-x64.exe`;
3. conecte o pendrive;
4. abra `blind-dev-setup-windows-x64.exe` para iniciar o assistente.

O executável abre um assistente que conduz a escolha do pendrive, o aceite das licenças, a confirmação do destino, a preparação e a verificação na mesma janela. Ao concluir, você pode iniciar o ambiente, abrir sua pasta, preparar outro pendrive ou sair.

Você também pode usar o [download direto do executável](https://github.com/woliveiras/blind-dev-setup/releases/latest/download/blind-dev-setup-windows-x64.exe), sem instalação. Se preferir baixar os guias e a licença junto com o programa, escolha `blind-dev-setup-windows-x64.zip`, extraia o ZIP e abra o mesmo executável dentro da pasta extraída.

### Download com curl

No PowerShell ou Prompt de Comando do Windows:

```powershell
curl.exe -fL "https://github.com/woliveiras/blind-dev-setup/releases/latest/download/blind-dev-setup-windows-x64.exe" -o "blind-dev-setup-windows-x64.exe"
```

### Download com wget

Se `wget` estiver instalado:

```powershell
wget --https-only -O "blind-dev-setup-windows-x64.exe" "https://github.com/woliveiras/blind-dev-setup/releases/latest/download/blind-dev-setup-windows-x64.exe"
```

O [ZIP mais recente](https://github.com/woliveiras/blind-dev-setup/releases/latest/download/blind-dev-setup-windows-x64.zip) e o arquivo `SHA256SUMS.txt` também ficam disponíveis em todas as releases.

O executável atual não possui assinatura digital. Por isso, o Windows pode apresentar um aviso de segurança mesmo quando o arquivo veio da release oficial. Confira se o endereço começa com `https://github.com/woliveiras/blind-dev-setup/` e, se quiser validar o download, compare seu SHA-256 com `SHA256SUMS.txt`.

## Uso

### Assistente na mesma janela

Conecte o pendrive e abra `blind-dev-setup-windows-x64.exe` diretamente, sem argumentos. Na pergunta inicial, pressione Enter ou `1` para usar a opção com leitor de tela/NVDA, `2` para o modo sem leitor de tela ou `0` para sair. Você não precisa conhecer comandos para escolher o modo.

Essa escolha ajusta a apresentação do assistente; não inicia nem encerra o leitor de tela. Para a opção `1`, abra o leitor que já estiver disponível no computador antes de iniciar o assistente. O NVDA Portable do pendrive estará disponível depois da preparação.

O assistente tem cinco etapas:

1. **Escolher pendrive:** confira nome, letra, capacidade e espaço livre. Escolha explicitamente a unidade, mesmo quando houver apenas uma.
2. **Revisar aplicativos e licenças:** leia o resumo e abra os detalhes quando desejar consultar a lista completa, versões e links. Aceite as licenças somente se concordar.
3. **Confirmar destino:** confira a unidade novamente e confirme a preparação. Nada é instalado sem aceite e confirmação.
4. **Preparar e verificar:** acompanhe as mensagens curtas de progresso. A saída técnica dos instaladores fica separada no log; a opção de detalhes mostra seu caminho depois da operação.
5. **Concluir:** após a verificação bem-sucedida, escolha **Iniciar ambiente** (`I` no modo com leitor de tela) para iniciar NVDA Portable e VS Code com o mesmo fluxo de `START.cmd`. Essa ação exige sua escolha explícita e mantém o menu disponível. As outras opções continuam sendo `1` para abrir a pasta no Explorador, `2` para outro pendrive, `D` para detalhes e `0` para sair. Abrir a pasta não inicia os aplicativos; nela, `START.cmd` também permite iniciá-los depois.

No modo com leitor de tela, as opções são anunciadas em texto e respondidas com Enter: `R` atualiza os destinos, `0` sai, `D` abre detalhes quando oferecido, `ACEITO` aceita as licenças e `SIM` confirma o destino. O modo sem leitor de tela usa formulários por teclado; siga as opções e teclas indicadas na tela. Os dois modos mantêm as mesmas confirmações e proteções.

Se a unidade já tiver uma instalação, o assistente oferece verificá-la, sem preparar novamente nem sobrescrever. Na conclusão ou em caso de falha, a opção de detalhes mostra um trecho do log e permite continuar na mesma janela. O resumo não mistura a saída extensa de downloads e instaladores com as perguntas.

Ao escolher **Iniciar ambiente**, o assistente confere novamente o destino e verifica o ambiente antes de executar `START.cmd`. Se essa etapa falhar, consulte `D` para ver os detalhes; o menu continua disponível. A mensagem de inicialização solicitada não comprova que as janelas abriram corretamente: confirme a abertura de NVDA e VS Code.

A conclusão confirma preparação e verificação estrutural. A avaliação de uso com NVDA ainda exige o protocolo prático; o nome “modo com leitor de tela” descreve o formato de interação, não uma aprovação empírica. As extensões Python permanecem opcionais e têm um guia próprio.

Para selecionar um modo diretamente ao iniciar pelo terminal, use `wizard --accessible` ou `wizard --visual`. Com entrada redirecionada, o assistente usa o modo com leitor de tela. Os comandos de automação abaixo continuam disponíveis.

### Linha de comando opcional

Os comandos abaixo permanecem disponíveis para automação e diagnóstico. Para usá-los, abra o PowerShell na pasta do executável pelo teclado:

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

Depois, abra E:\blind-dev-setup\START.cmd para iniciar NVDA e VS Code, ou DEV-SHELL.cmd para abrir o shell portátil. Os launchers de desenvolvimento preparam o mise no caminho atual e conferem suas configurações gerenciadas antes de autorizá-las. Se uma configuração tiver sido alterada, registre o erro em vez de contornar a conferência.

`verify` confere estrutura, executáveis esperados e configurações gerenciadas sem executar programas. Ele não prova a integridade completa de todos os arquivos extraídos nem substitui a execução dos exemplos e o teste com NVDA.

Se `list-targets` não encontrar nada, desconecte o pendrive, conecte-o novamente, aguarde alguns segundos e repita o comando. Se o Windows não atribuir uma letra ou o sistema de arquivos não for NTFS, a ferramenta explicará o problema sem tentar formatar a unidade.

## Validar uma revisão no Windows

O pacote de teste inclui `TESTE-WINDOWS.md`, `RESULTADOS-VALIDACAO.md` e `BUILD-INFO.txt`. Confira o commit completo, o estado das fontes (`clean` ou `modified`) e o SHA-256 antes do teste. Para uma revisão auditável, use o artefato da CI aprovada para aquele commit. Um pacote de teste pode incluir correções ainda ausentes na release pública mais recente.

Siga o [roteiro do notebook](internal/templates/files/docs/TESTE-WINDOWS.md) como usuário padrão e registre os resultados na [ficha de validação](internal/templates/files/docs/RESULTADOS-VALIDACAO.md). Ele cobre USB real, origem das ferramentas, launchers, exemplos, recusas, reconexão e mudança da letra. O script de volume virtual da CI não é necessário no notebook.

A edição assistida e a depuração Python exigem a escolha explícita das [extensões oficiais](internal/templates/files/docs/EXTENSOES-PYTHON.md). O [guia Python](internal/templates/files/docs/PYTHON-COM-UV.md) inclui um exemplo com breakpoint, diagnóstico e traceback intencional. As versões das extensões devem ser registradas; ainda não há uma combinação empiricamente aprovada pelo projeto.

## Compilar o projeto

Esta seção é somente para desenvolvimento. Usuários podem utilizar os downloads prontos descritos acima.

Requisitos para compilar:

- Go 1.27.4 ou posterior;
- Git;
- Windows 11 x64 para executar `prepare`.

No PowerShell:

```powershell
go test ./...
go build -trimpath -o blind-dev-setup-windows-x64.exe ./cmd/blind-dev-setup
```

Em macOS ou Linux é possível testar e gerar o executável Windows:

```bash
go test ./...
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -o blind-dev-setup-windows-x64.exe ./cmd/blind-dev-setup
```

O script `scripts/build.ps1` executa testes e análise estática, cria o executável e o ZIP em `dist` e gera `SHA256SUMS.txt`.

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

VS Code não recebe extensões automaticamente; o guia Python documenta a instalação opcional e o registro das versões. DBeaver precisa de validação empírica com NVDA. Atualização transacional, Windows arm64 e Linux estão no roadmap.

Leia [a decisão arquitetural](docs/adr/0001-windows-portable-builder.md), [a decisão sobre o assistente](docs/adr/0002-guided-wizard.md), [os limites das evidências](docs/research/evidence-boundaries.md), o [roadmap](ROADMAP.md), as [diretrizes de contribuição](CONTRIBUTING.md) e a [política de segurança](SECURITY.md).

## Licenças

O código deste projeto usa licença MIT. As ferramentas instaladas mantêm suas próprias licenças. Seus identificadores, links, versões, URLs e hashes estão em internal/bundle/windows-x64.json; os binários de terceiros não são armazenados no Git.
