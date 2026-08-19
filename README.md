# blind-dev-setup

`blind-dev-setup` cria, em um pendrive NTFS, um ambiente de desenvolvimento Windows pensado para pessoas cegas. A primeira versão inclui:

- Git for Windows, Visual Studio Code Portable, NVDA Portable e Notepad++ Portable;
- DBeaver Community em arquivo ZIP, com workspace no próprio pendrive;
- Python e Node.js instalados pelo mise em versões fixadas;
- uv para ambientes e projetos Python;
- pnpm para workspaces Node.js;
- inicializadores simples, lineares e compatíveis com leitores de tela.

O projeto está em desenvolvimento. A primeira versão suportará Windows 11 x64. Ela não formata o pendrive, não escolhe uma unidade automaticamente e não modifica uma instalação existente.

## Estado de acessibilidade

A presença de uma ferramenta no pendrive não prova que todos os fluxos dela são acessíveis. VS Code é o editor principal; Notepad++ é um editor auxiliar. O DBeaver é experimental até que seus fluxos principais sejam validados no Windows com NVDA por pessoas cegas.

Consulte [a decisão arquitetural](docs/adr/0001-windows-portable-builder.md), [os limites das evidências](docs/research/evidence-boundaries.md) e o [roadmap](ROADMAP.md).

## Licenças de terceiros

O executável deste projeto é MIT. As ferramentas instaladas mantêm suas próprias licenças. O gerador baixa cada artefato diretamente da fonte oficial e exige aceite explícito antes da preparação; ele não redistribui os binários de terceiros no repositório.

