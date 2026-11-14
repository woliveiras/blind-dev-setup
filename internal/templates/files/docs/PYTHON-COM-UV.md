# Python com uv e VS Code

Python e uv são instalados pelo mise em versões fixadas. Abra START.cmd para usar o terminal do VS Code ou DEV-SHELL.cmd para abrir um PowerShell. Os comandos abaixo são para esse terminal portátil.

## Executar o exemplo

    cd "$env:BLIND_DEV_ROOT\workspace\python-example"
    uv sync --python (mise which python) --no-python-downloads
    uv run --no-python-downloads python main.py
    uv run --no-python-downloads python debug_example.py

Resultados esperados: uma saudação e `Média: 4.0`. A seleção explícita do Python e `--no-python-downloads` evitam baixar outro interpretador para este teste.

Confira a origem do interpretador:

    uv run --no-python-downloads python -c "import sys; print(sys.executable); print(sys.base_prefix)"

Os caminhos devem apontar para o pendrive: `.venv` e a instalação em `data\mise`. Um Python do computador pode esconder uma falha do ambiente portátil.

`uv sync` cria o `.venv` e o `uv.lock`. Preserve o lock junto ao projeto. Para adicionar uma dependência, use `uv add nome-do-pacote`; em projetos com pytest, use `uv run pytest`. Não é necessário ativar o `.venv` manualmente para usar `uv run`.

## Extensões: escolha explícita

O gerador não instala extensões nem concede consentimento por você. Para executar a parte de edição e depuração do protocolo, leia [EXTENSOES-PYTHON.md](EXTENSOES-PYTHON.md) e instale o conjunto indicado. Sem ele, a execução pelo terminal continua disponível, mas os testes de símbolos, diagnósticos e depuração no editor ficam pendentes.

Depois de criar o `.venv`, abra `workspace\python-example\python-example.code-workspace` no VS Code portátil. Pelo teclado, use Ctrl+Shift+P, procure `Workspaces: Open Workspace from File` (o nome pode aparecer traduzido) e escolha esse arquivo. Confie no workspace apenas depois de revisar os exemplos fornecidos.

A configuração incluída aponta o depurador para `.venv\Scripts\python.exe`. Para a análise no editor, execute `Python: Select Interpreter` na paleta, escolha `Enter interpreter path` e indique esse mesmo arquivo se a extensão tiver selecionado outro interpretador. Isso é particularmente importante depois de trocar a letra da unidade.

## Exercício de diagnóstico, breakpoint e traceback

1. Abra `debug_example.py` com Ctrl+P. Use Ctrl+Shift+O para encontrar `average`.
2. Na linha `total += value`, coloque um breakpoint com F9.
3. Na área Executar e Depurar, escolha `Python: exemplo sem erro` e pressione F5. Confirme a parada, leia `value` e `total` em Variáveis, avance com F10 e continue com F5. Ao final, o terminal deve mostrar `Média: 4.0`.
4. Para um diagnóstico estático, remova temporariamente os dois-pontos de `for value in values:`. Salve, abra Problemas com Ctrl+Shift+M e leia o erro. Use Ctrl+Z e salve para restaurar o arquivo antes de executar novamente.
5. No terminal, execute `uv run --no-python-downloads python debug_example.py --fail`. O código de saída deve ser diferente de zero, com `ZeroDivisionError` na divisão de `average`. Esse erro é intencional: a lista está vazia.
6. Localize a linha do traceback no editor e execute novamente sem `--fail`. O resultado deve voltar a ser `Média: 4.0`. A configuração `Python: traceback intencional` permite repetir o cenário no depurador.

Registre separadamente falha de execução, dificuldade de navegação e ausência de anúncio do leitor de tela. Uma interface que abre não comprova acessibilidade.

## Mudança da letra da unidade

Um `.venv` pode guardar caminhos absolutos. Se deixar de funcionar após uma mudança de letra, feche os programas, renomeie a pasta `.venv` para um nome de cópia ainda inexistente e repita o comando `uv sync` acima. Confira o novo ambiente antes de remover a cópia. Reabra o arquivo de workspace e selecione o interpretador atual.

O código, `pyproject.toml` e `uv.lock` são os dados duráveis; `.venv` pode ser recriado.
