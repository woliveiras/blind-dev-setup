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

## A. Cinco etapas na mesma janela

Extraia o ZIP da revisão escolhida, conecte o pendrive e abra `blind-dev-setup-windows-x64.exe` diretamente, sem argumentos. O executável deve iniciar o assistente. Na pergunta inicial, use Enter ou `1` para a opção com leitor de tela/NVDA, `2` para iniciar sem leitor de tela ou `0` para sair. Registre se essa primeira escolha foi lida e pôde ser respondida sem outro terminal. Escolher a opção com leitor de tela não deve iniciar NVDA; a preparação usa o leitor que já estiver disponível no computador.

Faça o primeiro teste no modo com leitor de tela e registre os cinco passos:

1. **Escolher pendrive:** confira nome, letra, tamanho, espaço livre e NTFS. Digite o número da unidade conferida; mesmo uma unidade única exige escolha. `R` atualiza a lista e `0` sai. Um pendrive de 32 GB pode aparecer com aproximadamente 29,7 GiB; o requisito efetivo é ter 12 GiB livres.
2. **Revisar aplicativos e licenças:** confira o resumo. Use `D` para abrir os detalhes oferecidos e localizar versões e links de licença. Depois volte ao fluxo e digite `ACEITO` somente se concordar.
3. **Confirmar destino:** confira novamente a unidade e digite `SIM` somente para autorizar o destino escolhido.
4. **Preparar e verificar:** acompanhe mensagens curtas de progresso. A saída técnica dos instaladores deve ficar no log, não misturada às perguntas. A janela deve permanecer utilizável pelo leitor de tela, sem depender de cor ou animação. Depois da conclusão ou de uma falha, consulte `D` e confira o trecho e o caminho do log.
5. **Concluir:** depois da preparação e verificação bem-sucedidas, confira que **Iniciar ambiente** aparece em destaque. Digite `I` no modo com leitor de tela ou escolha essa ação no formulário para iniciar NVDA Portable e VS Code pelo fluxo de START.cmd. Antes dessa escolha explícita, os aplicativos não devem abrir. Confirme que o menu continua disponível após a abertura. As outras opções são `1` para abrir a pasta, `2` para outro destino, `D` para detalhes e `0` para sair. Teste `1` separadamente: abrir a pasta não deve iniciar aplicativos. Sair deve encerrar o assistente sem uma segunda pergunta para fechar a janela.

Esperado: nenhum arquivo pessoal alterado, pasta final `blind-dev-setup` na unidade selecionada, marcador `READY.txt` e verificação concluída antes do menu de sucesso. Em caso de falha, consulte os detalhes e preserve o log indicado, removendo informações pessoais antes de compartilhar. O assistente deve permitir continuar na mesma janela, sem apresentar uma falha como sucesso.

Repita a escolha de modo usando a opção sem leitor de tela, desta vez verificando a instalação já existente, sem nova etapa de licenças. Navegue pelos formulários usando as teclas indicadas e confirme que o texto dos controles e o foco permanecem compreensíveis. Registre esse resultado separadamente: o funcionamento do modo com leitor de tela não aprova o modo sem leitor de tela, nem o inverso.

A verificação é estrutural e de configurações; ela não substitui os testes de execução e acessibilidade a seguir nem prova a integridade completa dos arquivos de terceiros. O nome “modo com leitor de tela” não significa validação empírica com NVDA. As extensões Python continuam opcionais.

## B. Primeira abertura e origem das ferramentas

Antes de escolher Iniciar ambiente no menu final, feche outras janelas do VS Code para reduzir interferência de instâncias já abertas. Escolha `I` antes de abrir DEV-SHELL.cmd e confirme que NVDA Portable e VS Code abriram e que foco e mensagens podem ser percebidos. Volte à janela do assistente e confirme que o menu permanece disponível. Depois, feche os aplicativos e teste também `START.cmd` diretamente na pasta `blind-dev-setup` do pendrive; ele deve iniciar o mesmo ambiente. Se já houver NVDA em execução, registre qual instância está ativa; não conte a mera execução do launcher como validação do NVDA portátil.

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

Antes de autorizar uma preparação, teste estas interações no modo com leitor de tela; depois confira as ações equivalentes pelos controles do modo sem leitor de tela:

- Informe uma opção inválida na lista: a mensagem deve explicar o problema e permitir responder novamente na mesma janela.
- Use `R`: a lista deve ser atualizada sem fechar o programa.
- Abra `D` na revisão: os detalhes devem mostrar aplicativos/licenças e devolver o controle à revisão, sem iniciar instalação.
- Recuse o aceite das licenças e cancele a confirmação do destino, em tentativas separadas: nenhuma preparação pode começar.
- Use `0` na seleção: o programa deve sair normalmente, sem uma segunda pergunta para fechar a janela.

Depois da preparação bem-sucedida, abra o assistente novamente e selecione a mesma unidade. Ele deve explicar que o ambiente já existe e oferecer sua verificação, sem sobrescrever. Digite `SIM` para verificar a instalação existente: não deve haver nova preparação nem novo aceite de licenças. Confira que `READY.txt` e os arquivos de exemplo permanecem presentes. Uma verificação bem-sucedida deve levar ao menu final na mesma janela; uma falha deve manter a mensagem e permitir voltar à seleção, sem apresentar sucesso nem exigir outro terminal.

Não use o disco do sistema, não formate uma unidade para produzir erros e não exclua executáveis do ambiente para este teste manual; esses cenários pertencem aos testes automatizados.

Se houver uma falha de preparação, registre-a antes de escolher uma opção para tentar novamente. Não apague uma instalação anterior para contornar a recusa. Uma nova preparação exige destino sem instalação existente; o cache de downloads pode ser reutilizado. Se aparecer uma pasta temporária após interrupção brusca, preserve as evidências e investigue antes de removê-la.

### Diagnóstico opcional pela linha de comando

O fluxo principal acima não exige comandos. Para iniciar um modo diretamente pelo terminal, use `wizard --accessible` ou `wizard --visual`. Com entrada redirecionada, o assistente usa o modo com leitor de tela. Para conferir também a interface não interativa, abra o PowerShell na pasta do executável. Troque `E:\` pela letra conferida do pendrive e execute:

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

Feche VS Code, terminais, NVDA portátil e demais programas do pendrive. Ejete com segurança, reconecte e confira novamente a letra no assistente, usando `R` quando necessário. Selecione a unidade já preparada e digite `SIM` para verificar a instalação existente, sem sobrescrevê-la. Ao concluir, escolha `I` para iniciar o ambiente e repita os comandos de origem das ferramentas. Confira também `1` para abrir a pasta na nova letra. A verificação, o início do ambiente e a abertura da pasta devem ocorrer pelo próprio assistente, sem depender do comando opcional `verify`.

Se for possível testar outra letra com segurança, use uma letra livre para o mesmo pendrive pelas ferramentas do Windows, após fechar todos os programas. A alteração pode exigir ajuda administrativa; não execute os aplicativos como administrador. Não altere letras do sistema ou de outras unidades. Se não realizar essa etapa, marque mudança de letra como pendente.

Esperado: o ambiente base funciona com a letra nova. Para `.venv` e `node_modules`, siga os guias de recuperação, preservando código e lockfiles. Registre se a recriação foi necessária e se resolveu o problema.

## Resultado

Um teste passa quando o comportamento esperado e a leitura por teclado/NVDA forem confirmados nas etapas correspondentes. Marque `falhou` ou `não executado` quando apropriado; ausência de relato não é aprovação. Qualquer falha no preparo, origem das ferramentas, launchers ou acesso por leitor de tela impede declarar o fluxo principal validado.

Guarde a ficha e as versões com a revisão testada. Compartilhe mensagens de erro e passos para reprodução, sem credenciais, chaves, caminhos pessoais desnecessários ou identificação de participantes.
