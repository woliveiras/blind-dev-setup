# Evidências, recomendações e validação

## Fatos observados

- O NVDA oferece cópia portátil, mas ela não cobre login, telas UAC ou aplicativos elevados como uma instalação permanente.
- O modo portátil do VS Code usa uma pasta `data` ao lado do executável para configurações e extensões.
- Python depende de indentação; NVDA possui anúncio de indentação por fala, tons ou ambos.
- pnpm usa links no layout de `node_modules`; NTFS é o formato suportado por este projeto.
- O workspace do DBeaver precisa ser iniciado com um caminho no pendrive para evitar dados no perfil do Windows.

## Decisões do projeto

- Ativar `editor.accessibilitySupport`, quatro espaços e inserção de espaços no perfil inicial do VS Code.
- Não instalar add-ons do NVDA automaticamente. Add-ons executam código com os privilégios do NVDA e escolhas de fala são pessoais.
- Não importar uma configuração NVDA pessoal completa. O guia descreve perfis de editor, terminal e navegador como opções que cada pessoa aplica e testa.
- Tratar VS Code como editor principal, Notepad++ como auxiliar e DBeaver como experimental.

## Hipóteses que precisam ser validadas

- Todos os launchers e diálogos da preparação são compreensíveis com NVDA em Windows 11 x64.
- A versão fixada do Notepad++ não apresenta regressão de leitura nas funções usadas.
- Navegação, conexão, consulta e leitura de resultados do DBeaver são utilizáveis por teclado e NVDA.
- A troca de letra da unidade não impede que uv e pnpm recriem ambientes a partir dos lockfiles.

Configuração e testes automatizados verificam arquivos e comandos, não a experiência real de uma pessoa cega. A aceitação prática requer teste em Windows limpo como usuário padrão e execução do protocolo documentado por pessoas cegas. Um notebook já usado também fornece evidência útil, mas é necessário registrar ferramentas existentes e conferir que os comandos vêm do pendrive.

## Retomada em 14 de novembro de 2026

- A [CI do commit 25ae3c1](https://github.com/woliveiras/blind-dev-setup/actions/runs/43606211658) executou preparação real, verificação e os quatro toolchains em Windows Server 2025, com um volume NTFS virtual. Isso não comprova descoberta de USB físico, interfaces gráficas, execução como usuário padrão ou acessibilidade no Windows 11.
- Essa execução revelou um bloqueio transitório do Windows ao publicar a pasta. O gerador passou a repetir somente falhas de bloqueio conhecidas, durante no máximo 30 segundos, sem substituir uma instalação existente.
- A [CI de integração 77cc506](https://github.com/woliveiras/blind-dev-setup/actions/runs/43607297298) aprovou o inicializador compartilhado e o novo verify, com comandos pelo PATH isolado das ferramentas do runner. Executou exemplos Python/Node, mudou a letra de E: para Z:, rejeitou configuração mise alterada e detectou node.exe ausente. O ZIP, os checksums e o commit em BUILD-INFO.txt foram conferidos antes de disponibilizar o artefato.
- Nessa integração, a execução nativa revelou uma dependência de Get-FileHash no Windows PowerShell iniciado de outro PowerShell. O inicializador passou a calcular hashes diretamente pela API do .NET. O teste também passou a comparar identidade de arquivos, pois o Windows pode expandir nomes curtos de caminhos. Os cenários de regressão passaram no runner Windows.
- A execução de [Release Please](https://github.com/woliveiras/blind-dev-setup/actions/runs/43606211597) encontrou corretamente a tag v0.3.1 e considerou somente os dois commits posteriores. A [PR 5](https://github.com/woliveiras/blind-dev-setup/pull/5) não repetiu as mudanças antigas da PR 4, encerrada sem merge. Não foi possível estabelecer a causa histórica da falha da PR 4; os dados atuais não justificam forçar outra versão ou reescrever releases.
- Cada pacote de teste deve ser identificado pelo commit e pelos checksums dos seus arquivos. Uma execução verde anterior não valida alterações posteriores; consulte a CI da revisão exata do pacote.

O [roteiro de teste](https://github.com/woliveiras/blind-dev-setup/wiki/Teste-no-Windows) e a [ficha de validação](https://github.com/woliveiras/blind-dev-setup/wiki/Ficha-de-validacao) na Wiki orientam a validação humana e o registro de resultados. Os arquivos correspondentes no ambiente gerado contêm links para essas páginas. Nenhum resultado manual deve ser marcado como aprovado antes de sua execução.
