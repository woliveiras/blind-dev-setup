# Roadmap

O estado de uma release deve ser conferido pelo pacote, seu BUILD-INFO.txt e pela execução CI correspondente. A presença de uma funcionalidade no código não significa aprovação de acessibilidade.

## Implementado para Windows 11 x64

- Descoberta de volumes USB, destino explícito e preparação de ambiente novo em NTFS, sem formatação pelo gerador.
- Download de artefatos oficiais fixados, conferência de tamanho/SHA-256, montagem temporária e publicação da pasta final.
- Git, VS Code, NVDA, Notepad++, DBeaver e mise portáteis; Python, Node.js, uv e pnpm fixados.
- Verificação estrutural de executáveis e configurações gerenciadas, sem executar aplicativos.
- Inicialização compartilhada do ambiente nos launchers de desenvolvimento e conferência da configuração do mise antes de autorizar o caminho atual.
- Exemplo Python com configuração de depuração, instalação opcional de extensões e guias de teste e registro de resultados.
- Teste de preparação real em volume NTFS temporário na CI do Windows. O resultado vale para o commit executado; consulte os logs e o artefato correspondente.

## Critérios para aprovar a retomada

- [ ] CI aprovada para a revisão exata entregue, com prepare, verify, ferramentas no local publicado e limpeza.
- [ ] Pacote de teste com checksum, commit e estado das fontes identificados.
- [ ] Teste em notebook Windows 11 x64, como usuário padrão, usando pendrive real: descoberta, preparação, launchers independentes, exemplos e recusas sem alteração de dados.
- [ ] Reconexão e mudança da letra, incluindo recriação de ambientes derivados quando necessária.
- [ ] Versões das extensões Python registradas e diagnóstico/depuração aprovados na combinação testada.
- [ ] Protocolo completo por teclado e NVDA com pessoas cegas, com resultados registrados; repetir em Windows limpo para comprovar o cenário sem ferramentas previamente instaladas.
- [x] Investigação do histórico de releases repetidas: a execução de 14/11 reconheceu v0.3.1 e a PR 5 contém apenas mudanças posteriores. [Evidências e limites](docs/research/evidence-boundaries.md). Conferir novamente as notas antes de publicar a próxima versão.

Não marque estes itens por inferência. Registre links da CI e fichas de validação sem identificação de participantes. Falhas bloqueadoras devem ser corrigidas e os passos afetados repetidos antes de declarar aprovação.

## Evolução após a validação

- Atualização transacional que preserve dados e configurações, com recuperação em caso de falha.
- Distribuição opcional reproduzível de extensões, com versões, origem, hashes, licenças e compatibilidade validados.
- Avaliar uma alternativa ao DBeaver caso a validação com NVDA não seja satisfatória; até lá ele permanece experimental.
- Windows arm64 após validação dos componentes e dos fluxos reais.
- Linux com leitor de tela, terminal, editor, empacotamento e launchers próprios. O suporte não será declarado por simples tradução dos arquivos .cmd.
