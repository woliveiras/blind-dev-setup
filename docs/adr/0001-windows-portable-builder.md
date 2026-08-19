# ADR 0001: gerador portátil e seguro para Windows

- Status: aceito
- Data: 2026-08-19
- Escopo: v0.1.0

## Contexto

O projeto deve permitir que uma pessoa prepare um pendrive com ferramentas de desenvolvimento para uma pessoa cega. O conteúdo inclui aplicativos de terceiros grandes, com licenças e ciclos de versão independentes. Escrever em uma unidade removível também cria risco de perda de dados se a ferramenta formatar, selecionar ou sobrescrever o destino errado.

Uma ISO ou imagem bruta seria rígida, difícil de atualizar e perigosa para dados existentes. Redistribuir todos os aplicativos dentro de uma imagem também complicaria licenças e proveniência.

## Decisão

A v0.1.0 será uma CLI escrita em Go e distribuída como um único executável Windows x64. Ela terá três operações:

- `plan`: mostra o destino, espaço estimado, componentes, versões e ações sem escrever;
- `prepare`: prepara uma instalação nova somente após receber um destino explícito e `--accept-licenses`;
- `verify`: confere a estrutura, o manifesto instalado e os executáveis esperados sem modificar arquivos.

O gerador nunca formata ou particiona uma unidade, nunca seleciona um destino automaticamente, recusa o disco do sistema e recusa sobrescrever `blind-dev-setup` já existente. Os arquivos são montados em uma pasta temporária na mesma unidade; o diretório final só aparece depois que todas as etapas terminam.

Cada binário de terceiro é baixado de uma URL oficial, em versão fixada, e validado por SHA-256 antes de uso. Os binários não entram no Git. O manifesto registra versão, fonte, licença, tipo de instalação e arquivos esperados.

VS Code será o editor principal. Notepad++ será auxiliar. DBeaver será incluído como ferramenta experimental de banco de dados: sua presença e execução serão verificadas automaticamente, mas acessibilidade com NVDA precisa de validação humana.

O mise instalará versões fixadas de Python, Node.js, uv e pnpm. A documentação orientará `uv sync` e `uv run` para Python e workspaces pnpm para Node.js. Ambientes `.venv` e `node_modules` poderão ser recriados quando a letra da unidade mudar; arquivos de projeto e lockfiles serão os artefatos portáveis duráveis.

As mensagens da CLI serão texto simples, lineares, sem animação, barra de progresso ou dependência de cor. Os launchers usarão nomes descritivos e poderão ser operados pelo teclado e por leitor de tela.

## Alternativas consideradas

### ISO ou imagem bruta

Rejeitada para a primeira versão. Ela aumentaria o risco de sobrescrever unidades, dificultaria atualizações e misturaria a distribuição do projeto com binários de terceiros.

### Script PowerShell sem executável

Rejeitado como interface principal. Políticas de execução e diferenças de ambiente tornam a experiência inicial menos previsível. PowerShell continuará disponível dentro do ambiente preparado.

### Empacotar todos os binários no repositório

Rejeitado. O gerador preserva as licenças e fontes oficiais de cada ferramenta, reduz o tamanho do projeto e torna a proveniência auditável.

## Consequências

A preparação inicial exige internet e pode levar tempo. O pendrive precisa usar NTFS para suportar adequadamente os links do pnpm e os toolchains. O gerador não corrige automaticamente um sistema de arquivos incompatível.

O NVDA Portable não oferece leitura de login, UAC ou aplicativos elevados como uma instalação permanente. O projeto documentará a instalação permanente como alternativa quando esses fluxos forem necessários.

Windows é a única plataforma suportada na v0.1.0. Linux permanece no roadmap e terá decisões e testes próprios.
