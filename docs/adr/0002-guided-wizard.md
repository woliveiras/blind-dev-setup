# ADR 0002: assistente em cinco etapas com modos visual e acessível

- Status: aceito
- Data: 2026-11-14
- Escopo: primeiro uso em Windows 11 x64

## Contexto

A preparação precisa ser concluída na mesma janela, sem transcrever comandos para outro terminal. Uma sequência extensa de opções, licenças e saída técnica de instaladores dificulta acompanhar o passo atual. A interface também precisa oferecer uma alternativa linear ao formulário visual para leitores de tela.

## Decisão

Usar formulários huh em cinco etapas: escolher pendrive, revisar aplicativos/licenças, confirmar destino, preparar/verificar e concluir. A versão da biblioteca fica fixada no go.mod.

Abrir o executável sem argumentos inicia o assistente. INICIAR-AQUI.cmd chama essa entrada padrão. Antes dos formulários interativos, uma pergunta curta em texto oferece Enter/1 para o modo acessível para NVDA, 2 para o visual e 0 para sair. Essa escolha não depende de conhecer argumentos da CLI. `wizard --accessible` e `wizard --visual` permitem iniciar diretamente um modo; entrada redirecionada usa o acessível.

O modo acessível apresenta texto linear, sem depender de cor, mouse, spinner ou atualização da mesma linha. O modo visual organiza os campos e as opções por etapa com navegação por teclado. Os dois compartilham a mesma política de destino, consentimento, preparação e verificação. O formato acessível ainda precisa de validação empírica com pessoas cegas e NVDA.

## Fluxo e proteções

1. Listar unidades com nome, letra, tamanho, espaço livre e impedimentos. Exigir seleção explícita mesmo para uma unidade única; permitir atualização e saída.
2. Mostrar resumo dos aplicativos e oferecer a lista extensa com versões e links de licenças sob detalhes. Exigir aceite explícito; abrir detalhes não autoriza uma instalação.
3. Confirmar separadamente o destino. No modo acessível, os consentimentos continuam sendo `ACEITO` e `SIM`; EOF, cancelamento e resposta inválida não os substituem.
4. Reutilizar a política de destino e a preparação transacional existentes, sem formatar, particionar ou sobrescrever instalações. Revalidar o destino antes de executar. Mostrar progresso resumido e guardar saída técnica de instaladores em log separado. Executar a verificação estrutural depois da preparação.
5. Apresentar sucesso somente após verificar. Oferecer abrir a pasta no Explorador, escolher outra unidade, consultar detalhes/log e sair. Abrir a pasta não inicia aplicativos: o usuário pode abrir START.cmd quando desejar. Falhas devem manter resumo, detalhes e opções de recuperação na mesma janela.

Uma instalação existente oferece verificação com confirmação, sem nova preparação nem sobrescrita. Concluir sua verificação leva ao mesmo menu final. O caminho do log e o trecho consultável em detalhes devem ajudar o diagnóstico sem ocupar a tela principal com saída extensa.

O inicializador não adiciona pausa após uma saída normal. Se o executável não iniciar ou terminar com erro não tratado, conserva a mensagem com uma pausa antes de fechar. Os comandos list-targets, plan, prepare, verify, version e help continuam disponíveis para automação e diagnóstico.

## Validação e limites

A CI executa o binário real em `wizard --accessible` sobre volume NTFS temporário: cancelamento/EOF antes de cada consentimento, preparação nova, verificação existente e saída. Mantém as verificações de preservação de arquivos, origem/versão das ferramentas pelo PATH, exemplos, mudança de letra e rejeição de arquivos incompletos ou configurações alteradas.

Respostas redirecionadas testam funcionamento e consentimento. O inicializador real, a pergunta de modo, o layout visual, os detalhes, o foco e os anúncios com NVDA exigem o roteiro manual no notebook. Registre os resultados de cada modo separadamente.

Sucesso significa ambiente preparado e verificado estruturalmente; não aprova acessibilidade ou a integridade completa dos aplicativos. A instalação das extensões Python continua opcional e guiada pela documentação. Assinatura digital permanece fora desta mudança.
