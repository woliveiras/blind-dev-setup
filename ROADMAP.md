# Roadmap

## v0.1.0 - Windows 11 x64

- Gerar e verificar um ambiente novo em pendrive NTFS.
- Instalar artefatos oficiais fixados e conferir SHA-256 antes de extrair ou executar.
- Manter configurações, caches, workspaces e dados no pendrive.
- Oferecer launchers para NVDA, VS Code, Notepad++, DBeaver e shell de desenvolvimento.
- Documentar uv, pnpm, limitações do NVDA Portable e protocolo de validação com leitor de tela.

## Depois da v0.1.0

- Validar a experiência completa com pessoas cegas em uma instalação limpa do Windows.
- Implementar atualização transacional que preserve dados e configurações.
- Avaliar uma alternativa ao DBeaver se a validação com NVDA não for satisfatória.
- Adicionar Windows arm64 após validação dos componentes.
- Projetar e validar suporte a Linux, incluindo leitor de tela, terminal, editor, empacotamento e launchers próprios da plataforma.

Linux não será uma simples tradução dos arquivos `.cmd`: o suporte só será declarado quando os fluxos equivalentes tiverem sido testados na plataforma.

