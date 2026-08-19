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

Configuração e testes automatizados verificam arquivos e comandos, não a experiência real de uma pessoa cega. A aceitação da v0.1.0 requer teste em Windows limpo como usuário padrão e execução do protocolo documentado por pessoas cegas.
