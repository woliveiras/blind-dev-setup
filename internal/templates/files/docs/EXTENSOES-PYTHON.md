# Extensões Python: instalação opcional e registro

Este conjunto permite executar o protocolo de edição, diagnóstico e depuração. Sua instalação é uma escolha do usuário e requer internet. As extensões não fazem parte dos artefatos com hash fixado pelo manifesto do gerador.

## Conjunto mínimo para o protocolo

| Extensão | Identificador | Função |
| --- | --- | --- |
| [Python, Microsoft](https://marketplace.visualstudio.com/items?itemName=ms-python.python) | `ms-python.python` | Seleção do interpretador e integração Python |
| [Python Debugger, Microsoft](https://marketplace.visualstudio.com/items?itemName=ms-python.debugpy) | `ms-python.debugpy` | Breakpoints, passos e variáveis |
| [Pylance, Microsoft](https://marketplace.visualstudio.com/items?itemName=ms-python.vscode-pylance) | `ms-python.vscode-pylance` | Símbolos e diagnósticos de código |

A extensão Python pode trazer dependências opcionais, incluindo Python Environments. Confira e registre também essas extensões. O workspace do exemplo mantém `python.useEnvironmentsExtension` desativado para usar o fluxo explícito de seleção do interpretador; ele não desinstala dependências.

## Instalar pelo teclado

1. Abra o VS Code usando START-VSCODE.cmd do pendrive. Não use um atalho do computador.
2. Pressione Ctrl+Shift+X para abrir Extensões. Pesquise `@id:ms-python.python`.
3. Confira o identificador, o publicador Microsoft, a licença e a versão estável oferecida. Use Tab para chegar a Instalar e pressione Enter se concordar.
4. Repita a conferência para `@id:ms-python.debugpy` e `@id:ms-python.vscode-pylance`; elas podem já ter sido instaladas como dependências.
5. Abra o workspace do exemplo conforme PYTHON-COM-UV.md e registre as versões antes de iniciar os testes.

As recomendações do workspace não instalam extensões automaticamente. Use a instalação padrão, sem habilitar versões de pré-lançamento. As extensões devem ficar em `tools\vscode\data\extensions` no pendrive.

## Versões e repetição do teste

Ainda não existe uma combinação de versões destas extensões aprovada empiricamente pelo projeto. Não interprete a instalação como certificação de compatibilidade ou acessibilidade. O teste deve registrar a versão exata oferecida e as dependências presentes; a aprovação vale para essa combinação com as versões de Windows, VS Code, Python e NVDA registradas.

No DEV-SHELL, liste as versões usando o CLI do VS Code portátil:

    & "$env:BLIND_DEV_ROOT\tools\vscode\bin\code.cmd" --list-extensions --show-versions

Copie o resultado para RESULTADOS-VALIDACAO.md. No menu de cada extensão, desative sua atualização automática durante a sessão de validação. Confira a configuração efetiva no VS Code portátil.

Para repetir uma combinação registrada, use `Install Another Version` no menu da extensão e selecione a versão registrada, quando disponível. Qualquer atualização exige repetir os passos afetados e registrar as novas versões. Fixar versões e hashes para distribuição automatizada é uma decisão posterior à validação, não uma garantia oferecida por este guia.

Fontes oficiais: [depuração Python](https://code.visualstudio.com/docs/python/debugging), [edição Python](https://code.visualstudio.com/docs/python/editing) e [CLI do VS Code](https://code.visualstudio.com/docs/configure/command-line).
