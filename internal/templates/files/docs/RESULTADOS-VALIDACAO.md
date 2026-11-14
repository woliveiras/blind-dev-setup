# Registro de validação Windows

Copie este arquivo para registrar uma sessão. Preencha `passou`, `falhou` ou `não executado`; não deixe um campo vazio com significado de aprovação. Use somente dados de teste e não identifique participantes.

## Ambiente e pacote

- Data:
- Versão do gerador:
- Commit completo e estado das fontes de BUILD-INFO.txt:
- Link da execução CI/release:
- SHA-256 do ZIP e resultado da comparação:
- Windows (edição, versão/build e arquitetura):
- Usuário padrão, sem elevação nos aplicativos:
- Máquina limpa ou de uso cotidiano:
- Ferramentas já instaladas no computador:
- Pendrive (capacidade, NTFS, espaço livre; sem número de série):
- Letras usadas antes/depois:
- Versões de VS Code, Python, Node.js, uv, pnpm e Git:
- NVDA (versão, instalado/portátil, sintetizador e ajustes relevantes):
- Extensões e versões (`code.cmd --list-extensions --show-versions`):
- Downloads adicionais ou mudanças de configuração:

## Funcionamento

| Etapa | Resultado | Observado / evidência |
| --- | --- | --- |
| Descoberta do USB e plano | não executado | |
| Prepare e arquivos pessoais preservados | não executado | |
| Verify | não executado | |
| START.cmd como primeira entrada | não executado | |
| START-VSCODE.cmd independente | não executado | |
| DEV-SHELL.cmd independente | não executado | |
| Versões e caminhos vêm do pendrive | não executado | |
| uv sync e exemplo Python normal | não executado | |
| Diagnóstico de sintaxe | não executado | |
| Breakpoint, passos e variáveis | não executado | |
| Traceback intencional e recuperação | não executado | |
| pnpm install e pnpm start | não executado | |
| Recusa sem destino e sem aceite | não executado | |
| Recusa de instalação existente sem alteração | não executado | |
| Reconexão e nova abertura | não executado | |
| Mudança de letra e nova abertura | não executado | |
| Recuperação de .venv/node_modules, se necessária | não executado | |
| Notepad++ | não executado | |
| DBeaver (banco/driver de teste) | não executado | |

## Teclado e NVDA

| Passo do protocolo | Resultado | Foco, anúncio e dificuldade observados |
| --- | --- | --- |
| 1. Criar arquivo | não executado | |
| 2. Escrever e ler código | não executado | |
| 3. Identificar indentação | não executado | |
| 4. Navegar entre símbolos | não executado | |
| 5. Ler diagnóstico | não executado | |
| 6. Ler execução no terminal | não executado | |
| 7. Navegar pelo traceback | não executado | |
| 8. Depurar e ler variáveis | não executado | |
| 9. Rever diff | não executado | |
| 10. Fazer commit local | não executado | |

## Falhas e decisão

Para cada falha, registre: passo, resultado esperado, resultado observado, mensagem/código de erro, reprodução e eventual recuperação. Não cole segredos nos logs.

- Falhas bloqueadoras:
- Etapas não executadas e motivo:
- Limitações de acessibilidade:
- Reteste necessário:
- Resultado desta combinação de versões: pendente.

Este registro vale para o ambiente descrito. Automação, presença de arquivos e uma sessão com usuário sem deficiência visual não substituem validação com pessoas cegas.
