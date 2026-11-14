# Contribuição

## Antes de alterar

Leia o ADR, o manifesto e os limites de evidência. Não trate a presença de arquivos, testes ou configurações como prova de acessibilidade real.

## Compilar o projeto

Esta seção é somente para desenvolvimento. Usuários podem utilizar o [download pronto](https://github.com/woliveiras/blind-dev-setup/releases/latest/download/blind-dev-setup-windows-x64.exe).

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

## Fluxo local

    gofmt -w cmd internal
    go vet ./...
    go test -race ./...
    GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -o bin/blind-dev-setup-windows-x64.exe ./cmd/blind-dev-setup

Depois de alterar as dependências Go e concluir `go mod tidy`, regenere os avisos de licença com `python3 scripts/update-go-notices.py` e confira a alteração em `internal/notices/THIRD-PARTY-NOTICES.txt`. O pacote inclui os avisos das dependências junto à licença do projeto.

Use commits pequenos com Conventional Commits. Não inclua binários de terceiros, credenciais, perfis pessoais do NVDA ou dados de usuários.

## Teste completo de preparação na CI

O job Windows da CI executa o assistente do binário real com `wizard --accessible`, em um disco virtual temporário de 32 GiB, formatado em NTFS. O modo com leitor de tela de huh permite uma transcrição linear controlada pela entrada padrão; a interface sem leitor de tela e os anúncios de NVDA exigem teste manual. As respostas são enviadas pela entrada padrão do executável real: seleção do destino, aceite das licenças, confirmação da preparação e saída. O assistente executa a preparação e a verificação automática na mesma sessão. O teste exige a mensagem de conclusão, executa `verify` novamente e confere `READY.txt` e a remoção da área de staging. Os componentes reais são baixados e conferidos por hash, e o toolchain é instalado pelo mise.

Antes da instalação, o teste executa o programa sem argumentos com entrada redirecionada, que usa o modo com leitor de tela, e responde `0` para cancelar. Com `--accessible` explícito, também cancela na seleção, antes do aceite das licenças e antes da confirmação do destino. Encerra a entrada (EOF) antes de cada consentimento para comprovar que o fim da entrada não autoriza a instalação. Cada saída é verificada separadamente e não pode criar ambiente, staging ou cache. Um arquivo já existente no volume deve permanecer intacto tanto após o cancelamento quanto depois da preparação. Após a instalação, uma nova sessão do assistente seleciona o ambiente existente e confirma apenas sua verificação. O argumento `wizard --accessible --target` identifica explicitamente o volume de teste, que não é USB; a inspeção do destino e as confirmações continuam obrigatórias, sem opções que aceitem os prompts automaticamente.

Depois da publicação, o teste inicializa o ambiente pelo helper compartilhado `config/initialize-toolchain.cmd` e executa Python, Node.js, uv e pnpm diretamente pelo `PATH`, conferindo origem e versões. O `PATH` inicial contém apenas diretórios do Windows, para impedir que ferramentas instaladas no runner escondam dependências ausentes. O teste confere também a origem de Git pelo PATH, sua versão e a execução de um comando pelo Bash portátil. Isso exercita a extração silenciosa do Git e sua etapa de pós-instalação. A extração usa o helper oficial 7zr.exe do 7-Zip, baixado com versão, tamanho e SHA-256 fixados no manifesto antes do Git. A licença e a fonte correspondente ficam documentadas em [Wiki de fontes e licenças](https://github.com/woliveiras/blind-dev-setup/wiki/Fontes-e-licencas) e em docs/LICENSE-7ZIP.txt do ambiente gerado. Não se depende do tar.exe do Windows para extrair o pacote do Git. O teste muda a letra do volume e repete a verificação e a execução dos comandos.

O teste também altera temporariamente `mise.toml` para comprovar que o helper recusa uma configuração modificada antes de confiar nela, e retira `node.exe` de sua pasta para comprovar que `verify` detecta um toolchain incompleto. Ambos os arquivos são restaurados em `finally`, e os comandos são verificados novamente depois da restauração.

Na primeira letra, o smoke executa `uv sync` com o interpretador portátil explícito e downloads de Python desativados, roda `debug_example.py` e confere que o ambiente virtual usa o Python do pendrive. Também executa `pnpm install` e `pnpm start` no workspace Node. Após a mudança de letra, repete os comandos de versão; a recriação de `.venv` e a depuração gráfica continuam no roteiro manual.

Cada chamada do assistente recebe um diretório TEMP/TMP próprio dentro da área temporária do teste. Em caso de saída não zero, a CI mostra as últimas 120 linhas dos arquivos `blind-dev-setup-*.log` dessa chamada e preserva o código de saída original. Os diretórios anteriores de TEMP/TMP são restaurados, e a limpeza final remove também esses logs; o teste não consulta logs de sessões anteriores.

O teste roda em pushes para `main`, pull requests e acionamentos manuais de CI pelo GitHub Actions. Não usa cache persistente; indisponibilidade das fontes oficiais ou divergência de hashes faz o job falhar. A etapa tem limite de 35 minutos, dentro de um job de 45 minutos. O VHDX cresce conforme os dados são escritos e é desmontado e removido em `finally`, com uma etapa `always()` para repetir a limpeza após interrupções. Uma parada forçada do runner pode impedir a limpeza; o runner hospedado é descartável.

Para reproduzir em Windows, compile com `scripts/build.ps1` e execute em PowerShell 7 como administrador, escolhendo uma pasta temporária que ainda não exista:

```powershell
./scripts/test-prepare.ps1 -WorkDirectory "$env:TEMP\blind-dev-setup-e2e"
```

O script cria e formata apenas o disco associado ao seu próprio arquivo VHDX; não recebe uma unidade física de destino. Em caso de interrupção, repita com `-Cleanup` para remover os recursos temporários. O gerador continua sem formatar unidades.

Esse teste usa o Windows do runner e um disco virtual. Não valida detecção USB, interfaces gráficas, acessibilidade com NVDA ou a experiência em Windows 11 como usuário padrão; esses pontos continuam no protocolo de validação manual.

Depois que o teste de preparação e a limpeza passam, o job confere os checksums, compara o executável dentro do ZIP com o executável testado e exige que `BUILD-INFO.txt` registre o commit da execução e `Source state: clean`. Então disponibiliza o executável, o ZIP e `SHA256SUMS.txt` como artefato `blind-dev-setup-windows-x64-<commit>-<tentativa>`, com retenção de 14 dias. Na execução aprovada de CI, baixe esse artefato em Artifacts, extraia o arquivo baixado e confira o SHA-256 do ZIP com `Get-FileHash -Algorithm SHA256 .\blind-dev-setup-windows-x64.zip`, comparando com a linha correspondente em `SHA256SUMS.txt`. O ZIP interno é o pacote para testar no notebook. O artefato identifica a revisão testada e não publica uma release; não use pacotes de execuções com falha.

## Releases

Release Please acompanha os Conventional Commits em `main` e mantém uma PR de release com versão e changelog. Mesclar essa PR é a autorização humana para criar a tag e a GitHub Release.

Quando a release é criada, o workflow compila a revisão exata da tag em Windows, executa os testes, confere a versão embutida e publica somente estes artefatos:

- `blind-dev-setup-windows-x64.exe`;
- `blind-dev-setup-windows-x64.zip`;
- `SHA256SUMS.txt`.

Não envie binários manualmente. O repositório precisa permitir que GitHub Actions crie pull requests em Settings, Actions, General, Workflow permissions. A publicação usa somente o `GITHUB_TOKEN` temporário do próprio workflow.

## Atualização de componentes

Uma alteração de versão precisa incluir:

1. release estável na fonte oficial;
2. URL direta HTTPS;
3. tamanho exato do artefato;
4. SHA-256 conferido;
5. licença e link oficial;
6. arquivos esperados depois da instalação;
7. testes automatizados;
8. protocolo de validação no Windows com NVDA.

Não use latest em URLs ou versões. Uma atualização do DBeaver ou Notepad++ não deve ser declarada acessível apenas porque abre ou porque os testes de arquivos passam.

## Acessibilidade

A CLI não interativa e o modo com leitor de tela devem continuar lineares, sem spinners, dependência de cor ou atualização no mesmo lugar da tela. O assistente usa formulários huh e oferece a opção com leitor de tela/NVDA antes da interface sem leitor de tela, sem exigir argumentos. A interface sem leitor de tela pode renderizar formulários, mas toda operação deve ter equivalente pelo teclado e no modo com leitor de tela. Detalhes extensos de aplicativos/licenças ficam sob uma opção explícita; saída técnica de instaladores fica no log indicado ao usuário, separada do resumo de progresso.

Teste os dois modos: o modo com leitor de tela com respostas redirecionadas cobre funcionamento, EOF e consentimento; ele não comprova foco, anúncios ou navegação com um leitor de tela real. Mudanças de transcrição precisam atualizar scripts/test-prepare.ps1. Preferências de fala e add-ons do NVDA permanecem opt-in.

Mudanças no fluxo principal precisam seguir o [protocolo de validação com NVDA](https://github.com/woliveiras/blind-dev-setup/wiki/Validacao-com-NVDA) no ambiente gerado. Registre resultados e limitações observadas, sem identificar participantes.

## Documentação de uso

A [Wiki](https://github.com/woliveiras/blind-dev-setup/wiki) é a fonte dos manuais para usuários. Edite as páginas nela; evite copiar os guias completos para o README ou para os templates. Os arquivos em `internal/templates/files/docs` e o `LEIA-ME.txt` do pacote contêm links para a Wiki. Os avisos legais permanecem incluídos na distribuição.

Os nomes desses arquivos são preservados para que os links de pacotes e as mensagens de diagnóstico continuem funcionando. Os guias completos exigem internet ou uma cópia salva pelo usuário. A ficha de validação deve ser copiada da Wiki para preenchimento local.

Para editar a Wiki pelo terminal, clone seu repositório separado:

```bash
git clone git@github.com:woliveiras/blind-dev-setup.wiki.git
cd blind-dev-setup.wiki
```

Edite as páginas Markdown, confira os links e o menu `_Sidebar.md`, faça commit e publique com `git push`. Alterações na Wiki não publicam uma release do programa. Mantenha os downloads no formato `releases/latest/download/blind-dev-setup-windows-x64.exe`, sem fixar a versão.
