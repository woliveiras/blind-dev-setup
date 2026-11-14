# Limitações do ambiente

- Somente Windows 11 x64 é suportado.
- O destino deve ser NTFS, ter pelo menos 12 GiB livres e não estar no disco do sistema.
- A ferramenta prepara apenas ambientes novos; atualização preservando dados está no roadmap.
- A preparação exige internet.
- O repositório não redistribui binários de terceiros; eles são baixados das fontes oficiais.
- VS Code não recebe extensões automaticamente. Edição assistida e depuração Python exigem a instalação opcional descrita em EXTENSOES-PYTHON.md. As versões das extensões são registradas na validação; não estão fixadas no manifesto do gerador.
- NVDA Portable não substitui uma instalação permanente para login, UAC, telas seguras ou aplicativos elevados.
- DBeaver está incluído, mas sua acessibilidade ainda é experimental.
- Troca da letra da unidade pode exigir recriação de .venv e node_modules.
- Credenciais, chaves SSH, tokens e perfis pessoais não são copiados.
- Linux está no roadmap e não é suportado por estes launchers.
- verify confere estrutura e configurações gerenciadas sem executar programas. Ele não valida integralmente todos os binários extraídos nem substitui o teste prático.
- mise.toml e mise.lock são configurações gerenciadas; alterações são recusadas pelos launchers de desenvolvimento. Não desative essa conferência para contornar um erro.
- O teste no notebook de uso cotidiano não prova independência de ferramentas do computador: confira os caminhos e registre o ambiente existente.
