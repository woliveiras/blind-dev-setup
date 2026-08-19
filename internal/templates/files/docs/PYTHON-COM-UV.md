# Python com uv

Python e uv são instalados pelo mise em versões fixadas. Abra DEV-SHELL.cmd; não chame o Python instalado no computador.

## Exemplo incluído

No shell:

    cd python-example
    uv sync
    uv run python main.py

Para adicionar uma dependência:

    uv add nome-do-pacote

Para executar testes quando o projeto usa pytest:

    uv run pytest

uv sync cria e sincroniza o .venv do projeto. O arquivo uv.lock, quando presente, registra a resolução reproduzível. Não ative o ambiente manualmente para comandos rotineiros; uv run seleciona o ambiente correto.

## Mudança da letra da unidade

Um .venv pode guardar caminhos absolutos. Se ele deixar de funcionar depois que o Windows atribuir outra letra ao pendrive, renomeie a pasta .venv para .venv-old e execute uv sync. Confirme o novo ambiente antes de remover a cópia antiga.

O código, pyproject.toml e uv.lock são portáveis; o .venv é derivado e pode ser recriado.
