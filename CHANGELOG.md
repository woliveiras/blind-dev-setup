# Changelog

As mudanças relevantes deste projeto serão registradas neste arquivo.

## [0.3.1](https://github.com/woliveiras/blind-dev-setup/compare/v0.3.0...v0.3.1) (2026-08-29)


### Corrigido

* extract mise executable into expected directory ([1ab0d5f](https://github.com/woliveiras/blind-dev-setup/commit/1ab0d5fe0c19bd30934cd9732337da02daa2f2a5))

## [0.3.0](https://github.com/woliveiras/blind-dev-setup/compare/v0.2.0...v0.3.0) (2026-08-29)


### Adicionado

* add accessible USB target discovery ([d96b3fb](https://github.com/woliveiras/blind-dev-setup/commit/d96b3fb27ca45f0a245424255860a6b8fdac3310))


### Documentação

* document prebuilt utility downloads ([387b26e](https://github.com/woliveiras/blind-dev-setup/commit/387b26e18210f353c12c3cee7f3096259ba1f8db))
* guide first-time USB target selection ([9e33c96](https://github.com/woliveiras/blind-dev-setup/commit/9e33c966e9b74c64feb790cda7ac7803d5d11a48))


### Build

* package portable Windows release artifacts ([b7e025d](https://github.com/woliveiras/blind-dev-setup/commit/b7e025d2f89204c314f6dbaa4068317ab0f62180))


### Automação

* automate GitHub executable releases ([f25d4f0](https://github.com/woliveiras/blind-dev-setup/commit/f25d4f09309da2e415d7c3a37d4a773c63f461dd))

## [0.2.0](https://github.com/woliveiras/blind-dev-setup/compare/v0.1.0...v0.2.0) (2026-08-19)


### Adicionado

* add accessible USB target discovery ([d96b3fb](https://github.com/woliveiras/blind-dev-setup/commit/d96b3fb27ca45f0a245424255860a6b8fdac3310))


### Documentação

* document prebuilt utility downloads ([387b26e](https://github.com/woliveiras/blind-dev-setup/commit/387b26e18210f353c12c3cee7f3096259ba1f8db))
* guide first-time USB target selection ([9e33c96](https://github.com/woliveiras/blind-dev-setup/commit/9e33c966e9b74c64feb790cda7ac7803d5d11a48))


### Build

* package portable Windows release artifacts ([b7e025d](https://github.com/woliveiras/blind-dev-setup/commit/b7e025d2f89204c314f6dbaa4068317ab0f62180))


### Automação

* automate GitHub executable releases ([f25d4f0](https://github.com/woliveiras/blind-dev-setup/commit/f25d4f09309da2e415d7c3a37d4a773c63f461dd))

## 0.1.0 - 2026-08-19

### Adicionado

- CLI com list-targets, plan, prepare, verify e version.
- Descoberta somente leitura de pendrives USB, com diagnóstico e próximos passos em texto linear.
- Manifesto Windows x64 com versões, fontes, licenças, tamanhos e SHA-256 fixados.
- Preparação transacional sem formatação ou seleção automática de unidade.
- Git, VS Code, NVDA, Notepad++, DBeaver e mise portáteis.
- Python, Node.js, uv e pnpm instalados pelo mise.
- Launchers, configurações portáteis, exemplos e documentação acessível.
- Testes de manifesto, download, ZIP, destino, preparação, toolchain, templates e verificação.

### Limites conhecidos

- Validação empírica completa em Windows e com pessoas cegas ainda pendente.
- DBeaver experimental quanto à acessibilidade.
- Atualizações e Linux permanecem no roadmap.
