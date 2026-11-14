# Protocolo de validação com NVDA

Antes dos dez passos, execute TESTE-WINDOWS.md e prepare o exemplo conforme PYTHON-COM-UV.md e EXTENSOES-PYTHON.md. Abra python-example.code-workspace no VS Code portátil. Sem as extensões, símbolos, diagnósticos e depuração ficam não executados.

Execute os dez testes por teclado, sem depender do mouse. Registre versão do Windows, versão do NVDA, sintetizador, características da máquina sem identificar a pessoa, extensões e resultado de cada passo em uma cópia de RESULTADOS-VALIDACAO.md.

1. Criar arquivo: abra um projeto vazio, crie hello.py e confirme nome, aba e foco no editor.
2. Escrever Python: crie uma função com if, for e return; leia por caractere, palavra e linha.
3. Identificar indentação: navegue por um bloco com três níveis e identifique com segurança o início e fim de cada bloco.
4. Navegar entre símbolos: salte entre funções e classes pelo seletor de símbolos ou outline.
5. Ler erro: introduza um erro simples e confirme que o diagnóstico é anunciado, localizado e navegável.
6. Executar: rode o programa e confirme que a saída do terminal é legível.
7. Ler traceback: provoque uma exceção, localize a linha, volte ao código e corrija.
8. Depurar: use debug_example.py e a configuração Python: exemplo sem erro; coloque breakpoint em total += value, inicie a sessão, avance, leia variáveis e volte ao código sem perder o foco.
9. Rever diff: altere o arquivo e navegue por arquivo, hunk e linha modificada.
10. Fazer commit: em uma cópia descartável do exemplo, inicialize um repositório Git local, configure nome e e-mail de teste somente nesse repositório, faça stage, escreva a mensagem e confirme o commit por teclado sem foco perdido. Não publique esse repositório nem copie sua identidade pessoal.

## Preparar o teste Git local

Faça uma cópia do exemplo em `workspace\projects\git-validation` (use outro nome se já existir), sem copiar `.venv`. No DEV-SHELL, dentro dessa cópia, execute:

    git init
    git config --local user.name "Teste de validacao"
    git config --local user.email "teste@example.invalid"
    git add main.py debug_example.py pyproject.toml python-example.code-workspace
    git commit -m "Exemplo inicial para validacao"

Esses dados valem somente para esse repositório. Não configure identidade global nem adicione um remoto. Abra a cópia no VS Code, altere `main.py` e execute os passos de diff, stage e commit por teclado. Não adicione `.venv`, credenciais ou arquivos pessoais ao teste.

## Ferramentas auxiliares

Valide separadamente:

- Notepad++: abrir, editar, pesquisar e salvar sem regressão de leitura;
- DBeaver: criar conexão de teste, navegar no banco, editar consulta, executar, ler erros e percorrer resultados;
- DEV-SHELL: usar Git, uv sync, uv run, pnpm install e pnpm start;
- remoção segura: fechar todos os programas antes de ejetar o pendrive.

Automação verifica arquivos e comandos; ela não aprova acessibilidade. O fluxo principal só pode ser considerado empiricamente validado depois de este protocolo ser executado em Windows limpo, como usuário padrão, por pessoas cegas.
