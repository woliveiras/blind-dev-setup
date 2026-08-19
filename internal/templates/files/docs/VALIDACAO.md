# Protocolo de validação com NVDA

Execute os dez testes por teclado, sem depender do mouse. Registre versão do Windows, versão do NVDA, sintetizador, computador e resultado de cada passo.

1. Criar arquivo: abra um projeto vazio, crie hello.py e confirme nome, aba e foco no editor.
2. Escrever Python: crie uma função com if, for e return; leia por caractere, palavra e linha.
3. Identificar indentação: navegue por um bloco com três níveis e identifique com segurança o início e fim de cada bloco.
4. Navegar entre símbolos: salte entre funções e classes pelo seletor de símbolos ou outline.
5. Ler erro: introduza um erro simples e confirme que o diagnóstico é anunciado, localizado e navegável.
6. Executar: rode o programa e confirme que a saída do terminal é legível.
7. Ler traceback: provoque uma exceção, localize a linha, volte ao código e corrija.
8. Depurar: coloque breakpoint, inicie a sessão, avance, leia variáveis e volte ao código sem perder o foco.
9. Rever diff: altere o arquivo e navegue por arquivo, hunk e linha modificada.
10. Fazer commit: faça stage, escreva a mensagem e confirme o commit por teclado sem foco perdido.

## Ferramentas auxiliares

Valide separadamente:

- Notepad++: abrir, editar, pesquisar e salvar sem regressão de leitura;
- DBeaver: criar conexão de teste, navegar no banco, editar consulta, executar, ler erros e percorrer resultados;
- DEV-SHELL: usar Git, uv sync, uv run, pnpm install e pnpm start;
- remoção segura: fechar todos os programas antes de ejetar o pendrive.

Automação verifica arquivos e comandos; ela não aprova acessibilidade. A primeira versão só pode ser considerada empiricamente validada depois de este protocolo ser executado em Windows limpo, como usuário padrão, por pessoas cegas.
