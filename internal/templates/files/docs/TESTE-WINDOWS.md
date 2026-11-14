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

## A. Descoberta e preparação na mesma janela

1. Conecte o pendrive e abra INICIAR-AQUI.cmd. Abrir o executável sem argumentos também deve iniciar o assistente.
2. Confira nome, letra, tamanho, espaço livre e NTFS. A capacidade em GiB pode ser menor que o valor comercial em GB, por exemplo 29,7 GiB em um pendrive vendido como 32 GB; o requisito efetivo é ter pelo menos 12 GiB livres.
3. Confira que nenhuma unidade foi selecionada automaticamente, mesmo quando há apenas uma. Use `R` para atualizar a lista; depois digite o número correspondente ao pendrive conferido.
4. Leia o plano com os aplicativos, versões e links de licenças. Digite `ACEITO` somente se concordar com as licenças e `SIM` para confirmar o destino quando solicitado.
5. Acompanhe a preparação e a verificação até o fim, na mesma janela. Não é preciso abrir outro terminal ou copiar um comando para continuar.
6. Confira o menu de conclusão: `1` abre a pasta do ambiente no Explorador de Arquivos, `2` permite preparar outro pendrive e `0` sai. Escolha `1` e confirme que a pasta correta abriu, sem iniciar NVDA ou VS Code automaticamente. A opção `2` deve voltar à escolha explícita; `0` deve encerrar sem uma segunda pausa do inicializador.

Esperado: nenhum arquivo pessoal alterado, pasta final `blind-dev-setup` na unidade selecionada, marcador `READY.txt` e verificação concluída antes do menu de sucesso. Registre a saída integral em caso de falha, removendo nomes pessoais antes de compartilhá-la. Os erros e as opções para continuar devem permanecer na mesma janela; uma falha não pode aparecer como preparação concluída.

A verificação é estrutural e de configurações; ela não substitui os testes de execução e acessibilidade a seguir nem prova a integridade completa de todos os arquivos de terceiros. As extensões Python continuam opcionais e são instaladas pelo guia correspondente.

## B. Primeira abertura e origem das ferramentas

Na pasta aberta pelo assistente, localize `START.cmd`. Feche outras janelas do VS Code para reduzir interferência de instâncias já abertas e abra `START.cmd` primeiro, antes de DEV-SHELL.cmd. Se você saiu do assistente, navegue até a pasta `blind-dev-setup` do pendrive e abra o mesmo arquivo. Confirme que NVDA e VS Code abriram e que foco e mensagens podem ser percebidos. Se já houver NVDA em execução, registre qual instância está ativa; não conte a mera execução do launcher como validação do NVDA portátil.

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

## D. Cancelamento, recusa e recuperação na mesma janela

Antes de autorizar uma preparação, teste estas interações no assistente:

- Informe uma opção inválida na lista: a mensagem deve explicar o problema e permitir responder novamente na mesma janela.
- Use `R`: a lista deve ser atualizada sem fechar o programa.
- Recuse o aceite das licenças e cancele a confirmação do destino, em tentativas separadas: nenhuma preparação pode começar.
- Use `0` na seleção: o programa deve sair normalmente, sem uma segunda pergunta para fechar a janela.

Depois da preparação bem-sucedida, abra o assistente novamente e selecione a mesma unidade. Ele deve explicar que o ambiente já existe e oferecer sua verificação, sem sobrescrever. Digite `SIM` para verificar a instalação existente: não deve haver nova preparação nem novo aceite de licenças. Confira que `READY.txt` e os arquivos de exemplo permanecem presentes. Uma verificação bem-sucedida deve levar ao menu final na mesma janela; uma falha deve manter a mensagem e permitir voltar à seleção, sem apresentar sucesso nem exigir outro terminal.

Não use o disco do sistema, não formate uma unidade para produzir erros e não exclua executáveis do ambiente para este teste manual; esses cenários pertencem aos testes automatizados.

Se houver uma falha de preparação, registre-a antes de escolher uma opção para tentar novamente. Não apague uma instalação anterior para contornar a recusa. Uma nova preparação exige destino sem instalação existente; o cache de downloads pode ser reutilizado. Se aparecer uma pasta temporária após interrupção brusca, preserve as evidências e investigue antes de removê-la.

### Diagnóstico opcional pela linha de comando

O fluxo principal acima não exige comandos. Para conferir também a interface não interativa, abra o PowerShell na pasta do executável. Troque `E:\` pela letra conferida do pendrive e execute:

    .\blind-dev-setup-windows-x64.exe version
    .\blind-dev-setup-windows-x64.exe list-targets
    .\blind-dev-setup-windows-x64.exe plan --target E:\
    .\blind-dev-setup-windows-x64.exe verify --target E:\

Estes comandos de preparação devem retornar erro, sem alterar o destino:

    .\blind-dev-setup-windows-x64.exe prepare
    .\blind-dev-setup-windows-x64.exe prepare --target E:\
    .\blind-dev-setup-windows-x64.exe prepare --target E:\ --accept-licenses

Execute o último apenas quando a instalação já existir. Registre o resultado separadamente do teste do assistente.

## E. Reconexão e mudança da letra

Feche VS Code, terminais, NVDA portátil e demais programas do pendrive. Ejete com segurança, reconecte e confira novamente a letra no assistente, usando `R` quando necessário. Selecione a unidade já preparada e digite `SIM` para verificar a instalação existente, sem sobrescrevê-la. Ao concluir, use `1` para abrir sua pasta. Repita a primeira abertura pelo START.cmd e os comandos de origem das ferramentas. A verificação e a escolha da pasta devem ocorrer pelo próprio assistente, sem depender do comando opcional `verify`.

Se for possível testar outra letra com segurança, use uma letra livre para o mesmo pendrive pelas ferramentas do Windows, após fechar todos os programas. A alteração pode exigir ajuda administrativa; não execute os aplicativos como administrador. Não altere letras do sistema ou de outras unidades. Se não realizar essa etapa, marque mudança de letra como pendente.

Esperado: o ambiente base funciona com a letra nova. Para `.venv` e `node_modules`, siga os guias de recuperação, preservando código e lockfiles. Registre se a recriação foi necessária e se resolveu o problema.

## Resultado

Um teste passa quando o comportamento esperado e a leitura por teclado/NVDA forem confirmados nas etapas correspondentes. Marque `falhou` ou `não executado` quando apropriado; ausência de relato não é aprovação. Qualquer falha no preparo, origem das ferramentas, launchers ou acesso por leitor de tela impede declarar o fluxo principal validado.

Guarde a ficha e as versões com a revisão testada. Compartilhe mensagens de erro e passos para reprodução, sem credenciais, chaves, caminhos pessoais desnecessários ou identificação de participantes.
