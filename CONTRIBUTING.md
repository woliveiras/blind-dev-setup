# Contribuição

## Antes de alterar

Leia o ADR, o manifesto e os limites de evidência. Não trate a presença de arquivos, testes ou configurações como prova de acessibilidade real.

## Fluxo local

    gofmt -w cmd internal
    go vet ./...
    go test -race ./...
    GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -o bin/blind-dev-setup-windows-x64.exe ./cmd/blind-dev-setup

Use commits pequenos com Conventional Commits. Não inclua binários de terceiros, credenciais, perfis pessoais do NVDA ou dados de usuários.

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

Mensagens da CLI devem continuar lineares, sem spinners, dependência de cor ou atualização no mesmo lugar da tela. Toda operação deve ter alternativa por teclado. Preferências de fala e add-ons do NVDA permanecem opt-in.

Mudanças no fluxo principal precisam executar docs/VALIDACAO.md no ambiente gerado. Registre resultados e limitações observadas, sem identificar participantes.
