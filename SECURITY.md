# Segurança

## Versões suportadas

A release estável mais recente recebe correções de segurança.

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

Baixe o utilitário somente pela página oficial de releases. Cada release inclui `SHA256SUMS.txt` para conferir o executável e o ZIP. A versão atual ainda não possui assinatura digital de código; um checksum confirma a integridade do arquivo baixado, mas não substitui uma assinatura confiável.

Um pendrive continua sendo mídia removível. Proteja-o fisicamente, use criptografia quando houver dados sensíveis e não armazene tokens ou chaves sem uma política própria de proteção e recuperação.
