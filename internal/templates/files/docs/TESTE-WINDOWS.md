# Teste de aceitação no notebook Windows

Use este roteiro com o pacote da revisão que será avaliada. O ZIP de teste inclui este guia, RESULTADOS-VALIDACAO.md e BUILD-INFO.txt antes mesmo da preparação do pendrive. Preencha uma cópia de [RESULTADOS-VALIDACAO.md](RESULTADOS-VALIDACAO.md). Um resultado no seu notebook é evidência para essa máquina; se houver ferramentas já instaladas, registre-as. Isso não equivale a testar uma instalação limpa do Windows.

## Antes de começar

- Windows 11 x64 e usuário padrão. Não execute os aplicativos portáteis como administrador.
- Pendrive NTFS de pelo menos 32 GB, com 12 GiB livres. Use uma unidade sem uma pasta `blind-dev-setup` anterior; preserve uma cópia dos seus arquivos importantes. O programa não formata a unidade.
- Internet para a preparação, instalação opcional das extensões e primeiro uso de dependências/drivers.
- ZIP da revisão escolhida e seu `SHA256SUMS.txt`. Leia BUILD-INFO.txt e registre versão, commit completo, estado das fontes (clean/modified) e link da execução CI que produziu o pacote. Para uma revisão auditável, use um pacote de fontes clean cuja preparação completa tenha passado na CI. Compare `Get-FileHash -Algorithm SHA256 caminho-do-zip` com o arquivo de checksums antes de extrair.
- Registre Windows, arquitetura, VS Code, Python, Node.js e NVDA já instalados no computador. Não remova ferramentas pessoais para fazer o teste.
- Para operar a preparação com leitor de tela, use seu NVDA instalado ou outro leitor que você já utilize. O NVDA do pendrive só estará disponível depois da preparação.

Não execute `scripts/test-prepare.ps1` para este teste manual: esse script é destinado à CI e cria um disco virtual com privilégios administrativos.

## A. Descoberta e preparação

No PowerShell, na pasta do executável extraído:

    .\blind-dev-setup-windows-x64.exe version
    .\blind-dev-setup-windows-x64.exe list-targets

Confira nome, tamanho, letra e NTFS. Nos comandos seguintes, substitua `E:\` pela letra identificada. Não copie uma letra sem conferir a unidade.

    .\blind-dev-setup-windows-x64.exe plan --target E:\
    .\blind-dev-setup-windows-x64.exe prepare --target E:\ --accept-licenses
    .\blind-dev-setup-windows-x64.exe verify --target E:\

Aceite as licenças somente depois de consultá-las. Espere a conclusão de cada comando. Esperado: nenhum arquivo pessoal alterado, pasta final `E:\blind-dev-setup`, marcador `READY.txt` e `verify` concluído. Registre a saída integral em caso de falha, removendo nomes pessoais antes de compartilhá-la.

`verify` faz verificações estruturais e de configurações; ele não substitui os testes de execução e acessibilidade a seguir nem prova a integridade completa de todos os arquivos de terceiros.

## B. Primeira abertura e origem das ferramentas

Feche outras janelas do VS Code para reduzir interferência de instâncias já abertas. Abra `START.cmd` primeiro, antes de DEV-SHELL.cmd. Confirme que NVDA e VS Code abriram e que foco e mensagens podem ser percebidos. Se já houver NVDA em execução, registre qual instância está ativa; não conte a mera execução do launcher como validação do NVDA portátil.

Abra um novo terminal no VS Code usando a paleta Ctrl+Shift+P e o comando de criar terminal. Execute:

    python --version
    node --version
    uv --version
    pnpm --version
    git --version
    python -c "import sys; print(sys.executable)"
    node -p "process.execPath"
    mise which python
    mise which node
    mise which uv
    mise which pnpm
    where.exe git

Esperado: versões iguais às do manifesto e caminhos efetivos no pendrive. `where.exe` pode listar também ferramentas do computador, mas a primeira resolução de Git deve ser a portátil. Salve os resultados.

Feche os aplicativos. Repita a abertura com START-VSCODE.cmd e depois com DEV-SHELL.cmd, repetindo os comandos. Cada entrada deve funcionar sem depender de ter aberto outra antes. Se a inicialização recusar `mise.toml` ou `mise.lock` por alteração, registre o erro. Esses arquivos são gerenciados e conferidos antes de autorizar o mise; não remova a verificação nem aceite silenciosamente o arquivo alterado. Use um pacote confiável para reconstruir o ambiente, preservando seus projetos.

## C. Exemplos e acessibilidade

Os guias desta seção estão em `blind-dev-setup\docs` depois da preparação.

1. Siga `PYTHON-COM-UV.md`, incluindo criação do `.venv`, origem do Python e execução normal.
2. Se concordar, instale as extensões conforme `EXTENSOES-PYTHON.md`, registre suas versões e execute os exercícios de diagnóstico e depuração.
3. Execute os dez passos de `VALIDACAO.md` por teclado com NVDA. Registre separadamente funcionamento e leitura/foco.
4. No DEV-SHELL, entre em `workspace\node-workspace`, execute `pnpm install` e `pnpm start`. Confira a mensagem de exemplo e a origem de Node.js/pnpm.
5. Teste Notepad++ e DBeaver conforme VALIDACAO.md. Para DBeaver, use somente dados e credenciais de teste; registre o driver e qualquer download adicional. Se não houver banco de teste, marque essa parte como não executada.

O traceback com `--fail` no exemplo Python é intencional; sua presença e a possibilidade de localizar/corrigir a causa são o resultado esperado.

## D. Recusa de operações inválidas

Estes comandos devem terminar com erro e sem preparar ou substituir um ambiente:

    .\blind-dev-setup-windows-x64.exe prepare
    .\blind-dev-setup-windows-x64.exe prepare --target E:\
    .\blind-dev-setup-windows-x64.exe prepare --target E:\ --accept-licenses

Execute o último apenas depois da preparação bem-sucedida: ele deve recusar a pasta existente. Confira que o `READY.txt` e seus arquivos de exemplo continuam presentes. Não use o disco do sistema, não formate uma unidade para produzir erros e não exclua executáveis do ambiente para este teste manual; esses cenários pertencem aos testes automatizados.

Se houver uma falha de preparação, registre-a antes de tentar novamente. Não apague uma instalação anterior para contornar a recusa. O diretório de downloads pode ser reutilizado; a nova tentativa deve usar um destino sem instalação existente. Se aparecer uma pasta temporária após interrupção brusca, preserve as evidências e investigue antes de removê-la.

## E. Reconexão e mudança da letra

Feche VS Code, terminais, NVDA portátil e demais programas do pendrive. Ejete com segurança, reconecte e confirme novamente a letra em `list-targets`. Repita `verify`, a primeira abertura pelo START.cmd e os comandos de origem das ferramentas.

Se for possível testar outra letra com segurança, use uma letra livre para o mesmo pendrive pelas ferramentas do Windows, após fechar todos os programas. A alteração pode exigir ajuda administrativa; não execute os aplicativos como administrador. Não altere letras do sistema ou de outras unidades. Se não realizar essa etapa, marque mudança de letra como pendente.

Esperado: o ambiente base funciona com a letra nova. Para `.venv` e `node_modules`, siga os guias de recuperação, preservando código e lockfiles. Registre se a recriação foi necessária e se resolveu o problema.

## Resultado

Um teste passa quando o comportamento esperado e a leitura por teclado/NVDA forem confirmados nas etapas correspondentes. Marque `falhou` ou `não executado` quando apropriado; ausência de relato não é aprovação. Qualquer falha no preparo, origem das ferramentas, launchers ou acesso por leitor de tela impede declarar o fluxo principal validado.

Guarde a ficha e as versões com a revisão testada. Compartilhe mensagens de erro e passos para reprodução, sem credenciais, chaves, caminhos pessoais desnecessários ou identificação de participantes.
