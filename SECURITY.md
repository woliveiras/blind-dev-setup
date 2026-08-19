# Segurança

## Versões suportadas

A série 0.1.x recebe correções enquanto for a versão atual.

## Relatar vulnerabilidade

Use um GitHub Security Advisory privado em:

https://github.com/woliveiras/blind-dev-setup/security/advisories/new

Não abra uma issue pública antes de existir uma correção quando o problema puder causar execução de código, download adulterado, exposição de credenciais ou escrita fora da pasta de destino.

## Modelo de segurança

O projeto considera especialmente:

- escolha errada ou destrutiva da unidade;
- artefatos adulterados ou redirecionamentos inseguros;
- ZIP path traversal, links e expansão excessiva;
- execução de comandos por concatenação de shell;
- configurações mise não confiáveis;
- credenciais e perfis pessoais no pendrive;
- limitações do NVDA Portable em processos elevados.

O gerador não formata unidades, não sobrescreve instalações, verifica artefatos antes de executar ou extrair e passa argumentos diretamente aos processos. O diretório final só é publicado depois da conclusão.

Um pendrive continua sendo mídia removível. Proteja-o fisicamente, use criptografia quando houver dados sensíveis e não armazene tokens ou chaves sem uma política própria de proteção e recuperação.
