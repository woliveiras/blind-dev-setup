# Configuração do NVDA para programação

O projeto entrega uma cópia portátil limpa. Preferências de voz, velocidade, pronúncia e add-ons não são aplicadas automaticamente porque são pessoais e add-ons executam código dentro do NVDA.

## Limites da cópia portátil

NVDA Portable não oferece inicialização no login, leitura de telas UAC, leitura de telas seguras nem acesso confiável a aplicativos elevados. Evite executar VS Code, DBeaver ou terminais como administrador. Quando esses fluxos forem necessários, instale o NVDA permanentemente no computador pela fonte oficial.

## Perfil inicial sugerido

Estas são recomendações para testar, não valores universais:

- no perfil do editor, nível de símbolos All;
- anúncio de indentação por Speech para iniciantes;
- ignorar linhas vazias no anúncio de indentação;
- números de linha ativados;
- caracteres digitados apenas em controles de edição;
- palavras digitadas desativadas;
- teclas de comando ativadas durante o aprendizado;
- no terminal e navegador, nível de símbolos Most;
- fala de senhas em terminais aprimorados desativada.

Crie perfis separados para editor, terminal e navegador em Preferences, Configuration Profiles. Use o gatilho da aplicação atual e valide cada perfil.

Para Python, confira a pronúncia de parênteses, colchetes, chaves, dois-pontos, aspas, igual, maior, menor, asterisco, barras, underscore, hífen, arroba e cardinal. Sequências como ==, !=, <=, >=, //, ** e -> podem receber pronúncias próprias.

## VS Code

O perfil já ativa suporte a acessibilidade, quatro espaços, inserção de espaços e minimap desativado. Use:

- Alt+F1 para ajuda de acessibilidade;
- Alt+F2 para a vista acessível, inclusive em saída longa do terminal;
- Ctrl+Shift+M para o painel de problemas;
- Ctrl+Shift+O para símbolos do arquivo;
- F12 para ir à definição;
- Shift+F12 para referências.

## Add-ons

Não instale um add-on apenas porque ele aparece em uma lista. Confirme autor, versão compatível, origem e manutenção. IndentNav pode ajudar a navegar por níveis de indentação, mas deve ser opt-in e testado com a versão atual do NVDA.
