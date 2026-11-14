# Fontes e licenças

As versões, URLs e SHA-256 usados na preparação estão em config/manifest.json.

Documentação oficial usada para o desenho:

- VS Code Portable: https://code.visualstudio.com/docs/setup/portable
- Acessibilidade no VS Code: https://code.visualstudio.com/docs/editor/accessibility
- Licença dos binários VS Code: https://code.visualstudio.com/License/
- Guia do NVDA: https://download.nvaccess.org/documentation/en/userGuide.html
- Instalação do mise: https://mise.jdx.dev/installing-mise.html
- Ambientes uv: https://docs.astral.sh/uv/pip/environments/
- Projetos uv: https://docs.astral.sh/uv/concepts/projects/layout/
- Workspaces pnpm: https://pnpm.io/workspaces
- Estrutura de links pnpm: https://pnpm.io/symlinked-node-modules-structure
- DBeaver Community: https://dbeaver.io/download/

Cada ferramenta mantém sua própria licença. plan mostra o identificador e o manifesto contém o link correspondente. O aceite dado ao gerador não altera nem substitui os termos dos fornecedores.

## Helper 7-Zip reduzido (7zr)

A preparação baixa o executável de console `7zr.exe` oficial, versão 26.06, para extrair o pacote Portable Git sem abrir uma janela do autoextrator. O manifesto registra URL, tamanho e SHA-256; o helper fica em `tools/7zip/7zr.exe`.

Este helper reduzido pertence ao LZMA SDK, que Igor Pavlov colocou em domínio público. A licença genérica do pacote completo do 7-Zip não descreve corretamente esse binário específico. O `bin/7zr.exe` do SDK 26.06 foi comparado ao hash e tamanho fixados no manifesto: são os mesmos bytes.

- [Declaração oficial de domínio público e descrição de 7zr no LZMA SDK](https://www.7-zip.org/sdk.html).
- [SDK 26.06 com código-fonte, binários e DOC/lzma-sdk.txt](https://github.com/ip7z/7zip/releases/download/26.06/lzma2606.7z).
- [Release oficial 26.06](https://github.com/ip7z/7zip/releases/tag/26.06).
- [Alvo de compilação Alone7z que gera 7zr.exe](https://github.com/ip7z/7zip/blob/26.06/CPP/7zip/Bundles/Alone7z/makefile).
- [Aviso do helper e seção de licença do SDK incluídos neste ambiente](LICENSE-7ZIP.txt).

O projeto baixa o helper sem modificá-lo. O aviso incluído reproduz a seção LICENSE do DOC/lzma-sdk.txt da versão 26.06, com atribuição e links para o código-fonte correspondente.
