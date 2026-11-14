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

## Helper 7-Zip

A preparação baixa o executável de console `7zr.exe` oficial, versão 26.06, para extrair o pacote Portable Git sem abrir uma janela do autoextrator. O manifesto registra URL, tamanho e SHA-256; o helper fica em `tools/7zip/7zr.exe`.

- [Release oficial 26.06 e código-fonte correspondente](https://github.com/ip7z/7zip/releases/tag/26.06).
- [Código-fonte na tag 26.06](https://github.com/ip7z/7zip/tree/26.06).
- [Informações de licença oficiais](https://www.7-zip.org/license.txt), reproduzidas em [LICENSE-7ZIP.txt](LICENSE-7ZIP.txt).

O executável `7zr.exe` usa GNU LGPL 2.1 ou posterior, conforme as informações do fornecedor para os arquivos que não são `7z.dll`. Ele é distribuído sem alterações pelo fornecedor e baixado pelo gerador; o projeto não modifica seu código. Consulte o texto de licença e o código-fonte correspondente na release indicada.
