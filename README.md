# blind-dev-setup

Ambiente de desenvolvimento portátil para pessoas cegas, com assistente por teclado para Windows 11 x64, com ou sem leitor de tela.

**[Baixar para Windows 11 x64 (.exe)](https://github.com/woliveiras/blind-dev-setup/releases/latest/download/blind-dev-setup-windows-x64.exe)**

**[Abrir o guia de uso na Wiki](https://github.com/woliveiras/blind-dev-setup/wiki)**

## Baixar e usar

1. [Baixe o executável](https://github.com/woliveiras/blind-dev-setup/releases/latest/download/blind-dev-setup-windows-x64.exe). O link entrega o arquivo da release mais recente.
2. Conecte um pendrive NTFS de pelo menos 32 GB, com 12 GiB livres.
3. Abra `blind-dev-setup-windows-x64.exe` diretamente, sem argumentos, e siga o assistente.

Você não precisa instalar Go, Git nem compilar o projeto. A preparação exige internet. O programa não formata a unidade nem sobrescreve uma instalação existente.

Também estão disponíveis o [ZIP com o programa, links dos guias e licenças](https://github.com/woliveiras/blind-dev-setup/releases/latest/download/blind-dev-setup-windows-x64.zip) e o [arquivo SHA256SUMS.txt](https://github.com/woliveiras/blind-dev-setup/releases/latest/download/SHA256SUMS.txt). O executável não possui assinatura digital; consulte as [orientações de download](https://github.com/woliveiras/blind-dev-setup/wiki/Primeiros-passos).

## Guia de uso

A documentação para usuários fica na Wiki:

- [Primeiros passos e preparação do pendrive](https://github.com/woliveiras/blind-dev-setup/wiki/Primeiros-passos).
- [Uso diário e abertura das ferramentas](https://github.com/woliveiras/blind-dev-setup/wiki/Uso-diario).
- [NVDA e acessibilidade](https://github.com/woliveiras/blind-dev-setup/wiki/NVDA-e-acessibilidade).
- [Python com uv](https://github.com/woliveiras/blind-dev-setup/wiki/Python-com-uv) e [extensões Python](https://github.com/woliveiras/blind-dev-setup/wiki/Extensoes-Python).
- [Node.js com pnpm](https://github.com/woliveiras/blind-dev-setup/wiki/Nodejs-com-pnpm).
- [Solução de problemas](https://github.com/woliveiras/blind-dev-setup/wiki/Solucao-de-problemas) e [linha de comando](https://github.com/woliveiras/blind-dev-setup/wiki/Linha-de-comando).
- [Teste no Windows](https://github.com/woliveiras/blind-dev-setup/wiki/Teste-no-Windows) e [validação com NVDA](https://github.com/woliveiras/blind-dev-setup/wiki/Validacao-com-NVDA).

Os arquivos de ajuda no pacote e no pendrive apontam para a Wiki. Salve uma cópia das páginas se precisar consultar os guias completos sem internet.

## Ferramentas incluídas

Git Portable, VS Code Portable, NVDA Portable, Notepad++, DBeaver Community e mise, com Python, Node.js, uv e pnpm em versões fixadas. Configurações, caches, exemplos e projetos ficam no pendrive.

A acessibilidade do DBeaver ainda é experimental. A verificação automatizada confere arquivos e configurações; o uso com NVDA exige validação prática com pessoas cegas. Consulte as [limitações](https://github.com/woliveiras/blind-dev-setup/wiki/Limitacoes).

## Desenvolvimento

Para compilar, testar ou contribuir, leia [CONTRIBUTING.md](CONTRIBUTING.md). As [decisões arquiteturais](docs/adr/0001-windows-portable-builder.md), a [decisão sobre o assistente](docs/adr/0002-guided-wizard.md), os [limites das evidências](docs/research/evidence-boundaries.md), o [roadmap](ROADMAP.md) e a [política de segurança](SECURITY.md) ficam no repositório.

## Licenças

O código do projeto usa a [licença MIT](LICENSE). As ferramentas mantêm suas próprias licenças; consulte [fontes e licenças](https://github.com/woliveiras/blind-dev-setup/wiki/Fontes-e-licencas). Versões, URLs e hashes estão no [manifesto](internal/bundle/windows-x64.json). Os binários de terceiros não são armazenados no Git.
