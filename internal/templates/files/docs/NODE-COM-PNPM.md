# Node.js com pnpm

Node.js e pnpm são instalados pelo mise em versões fixadas. Abra DEV-SHELL.cmd.

## Exemplo incluído

No shell:

    cd node-workspace
    pnpm install
    pnpm start

O arquivo pnpm-workspace.yaml define os pacotes do workspace. Use o protocolo workspace: ao declarar dependências internas que não podem ser resolvidas no registro.

Mantenha pnpm-lock.yaml versionado. Ele permite reconstruir as dependências com:

    pnpm install --frozen-lockfile

## Portabilidade

pnpm economiza espaço usando links entre node_modules e seu store. Por isso esta versão exige NTFS. Se node_modules deixar de funcionar depois de uma mudança de letra, renomeie a pasta para node_modules-old, execute pnpm install e confirme o resultado antes de remover a cópia antiga.

O store compartilhado fica em data/pnpm/store no pendrive. Código, package.json, pnpm-workspace.yaml e pnpm-lock.yaml são os dados duráveis.
