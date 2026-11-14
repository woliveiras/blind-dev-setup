# ADR 0002: assistente de preparação na mesma janela

- Status: aceito
- Data: 2026-11-14
- Escopo: primeiro uso em Windows 11 x64

## Contexto

O inicializador anterior executava apenas list-targets e depois pedia para fechar a janela. Para preparar o pendrive, a pessoa precisava abrir outro terminal e transcrever comandos. Isso interrompia o fluxo iniciado pelo pacote e tornava a descoberta da unidade um ponto de bloqueio para usuários sem familiaridade com linha de comando.

## Decisão

Abrir o executável sem argumentos inicia um assistente de console. INICIAR-AQUI.cmd chama essa entrada padrão. Os comandos list-targets, plan, prepare, verify, version e help continuam disponíveis para automação e diagnóstico.

O assistente mantém seleção, plano, consentimento, preparação, verificação, recuperação de erros e conclusão na mesma janela. Ele usa mensagens lineares e respostas pelo teclado, sem exigir cor, mouse, spinner ou atualização da mesma linha.

O fluxo deve:

1. Listar unidades com nome, letra, tamanho, espaço livre e impedimentos.
2. Exigir seleção explícita mesmo quando houver uma única unidade; permitir R para atualizar e 0 para sair.
3. Apresentar componentes, versões e links de licença antes do aceite.
4. Exigir `ACEITO` para o aceite explícito das licenças e `SIM` para a confirmação do destino antes de iniciar a preparação.
5. Reutilizar a política de destino e a preparação transacional existentes, sem formatar, particionar ou sobrescrever instalações.
6. Executar a verificação estrutural depois de preparar e só então apresentar sucesso. Se o destino já tiver uma instalação, oferecer sua verificação com `SIM`, sem nova preparação ou sobrescrita, e apresentar o menu final somente após a verificação bem-sucedida.
7. Apresentar no fim as opções `1` para abrir a pasta do ambiente no Explorador de Arquivos, `2` para preparar outro pendrive e `0` para sair. A abertura da pasta não inicia aplicativos; o usuário pode abrir START.cmd quando desejar. Falhas devem manter mensagens e opções de recuperação na mesma janela.

O inicializador não adiciona uma pausa após a saída normal, pois o assistente já controla o encerramento. Se o executável não iniciar ou terminar com erro não tratado, o inicializador conserva a mensagem com uma pausa antes de fechar.

## Consequências e validação

O usuário pode concluir a preparação a partir do pacote sem transcrever comandos para outro terminal. Automação continua usando argumentos explícitos e não deve depender do assistente.

Testes devem cobrir respostas inválidas, atualização da lista, cancelamento, nenhuma unidade, uma ou várias unidades, verificação sem sobrescrever instalação existente, falha de preparação/verificação e as escolhas de conclusão. O teste no Windows deve verificar o inicializador real e os anúncios/foco com NVDA.

Sucesso no assistente significa ambiente preparado e verificado estruturalmente. Não é aprovação de acessibilidade nem validação de integridade completa dos aplicativos. A instalação das extensões Python permanece opcional e guiada pela documentação; o protocolo empírico continua necessário.
