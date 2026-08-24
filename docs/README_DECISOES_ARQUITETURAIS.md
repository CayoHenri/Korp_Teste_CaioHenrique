# Decisões arquiteturais

Este documento registra escolhas implementadas e seus efeitos. Os registros
usam um formato ADR simplificado: contexto, decisão e consequências.

## ADR-001 — Dois microsserviços em monorepo

**Status:** aceita.

**Contexto:** Estoque e Faturamento possuem regras e dados distintos, mas a
entrega e a avaliação ocorrem como um único projeto.

**Decisão:** manter dois módulos Go independentes em `services/estoque` e
`services/faturamento`, no mesmo repositório.

**Consequências:** responsabilidades e dependências ficam isoladas; execução e
documentação permanecem simples. A integração e a consistência eventual elevam
a complexidade em relação a um monólito.

## ADR-002 — Go, Gin e Clean Architecture pragmática

**Status:** aceita.

**Decisão:** usar Go com Gin na borda HTTP e organizar cada serviço em domain,
application, infrastructure e presentation.

**Consequências:** domínio e use cases não dependem do framework. A estrutura
cria mais arquivos, mas cada responsabilidade tem localização previsível. Não
são criadas abstrações que não protejam uma fronteira real.

## ADR-003 — PostgreSQL compartilhado com schemas exclusivos

**Status:** aceita para o escopo local.

**Decisão:** usar uma instância e dois schemas, `estoque` e `faturamento`.

**Restrição:** acesso cruzado às tabelas não é permitido. Integração ocorre por
HTTP ou evento.

**Consequências:** menor custo operacional com isolamento lógico. Não existe
isolamento físico; uma evolução pode mover os schemas para bancos diferentes.

## ADR-004 — GORM sem AutoMigrate

**Status:** aceita.

**Decisão:** GORM implementa repositórios e transações. Models ficam na
infraestrutura e são convertidos explicitamente para entidades.

**Consequências:** reduz código SQL repetitivo sem acoplar o domínio. Alterações
de schema não podem acontecer implicitamente pela API.

## ADR-005 — Migrations SQL com golang-migrate

**Status:** aceita.

**Decisão:** cada módulo mantém pares `up` e `down` e um comando próprio de
migration. No Docker, migrations rodam em containers de execução única.

**Consequências:** histórico auditável e startup da API sem alteração de schema.
O autor precisa manter reversões coerentes e controlar estados `dirty`.

## ADR-006 — Configuração explícita, sem fallback

**Status:** aceita.

**Decisão:** toda variável necessária é validada. Ausência ou formato inválido
interrompe o processo. `.env.example` documenta o contrato.

**Consequências:** erros de ambiente aparecem cedo e não há conexão silenciosa
com recursos incorretos. A execução exige preparação do `.env`.

## ADR-007 — HTTP para consulta e RabbitMQ para fechamento

**Status:** aceita.

**Decisão:** HTTP atende operações que precisam de resposta imediata. RabbitMQ
transporta solicitação e resultado da baixa.

**Consequências:** criação de nota falha rapidamente se o Estoque não puder
validar produtos. Depois que o fechamento é aceito, Estoque e Faturamento não
precisam estar disponíveis ao mesmo tempo.

## ADR-008 — Estado PROCESSANDO

**Status:** aceita.

**Contexto:** existe tempo entre solicitar baixa e receber o resultado.

**Decisão:** usar `ABERTA`, `PROCESSANDO` e `FECHADA`.

**Consequências:** o sistema representa a consistência eventual e impede edição
ou nova solicitação durante o processamento. O cliente precisa acompanhar o
resultado depois da resposta inicial.

## ADR-009 — Transactional Outbox

**Status:** aceita.

**Decisão:** status `PROCESSANDO` e evento são persistidos na mesma transação. Um
worker publica registros pendentes e aguarda confirmação do RabbitMQ.

**Consequências:** evita nota processando sem solicitação persistida. Ainda pode
haver publicação duplicada entre a confirmação do broker e `published_at`, por
isso idempotência continua obrigatória.

## ADR-010 — Entrega pelo menos uma vez e idempotência

**Status:** aceita.

**Decisão:** consumidores registram identificadores processados na mesma
transação do efeito de negócio.

**Consequências:** reentrega não duplica baixa ou transição. Há armazenamento e
lógica adicionais, mas não se promete “exactly once”, que seria inadequado para
esse fluxo distribuído.

## ADR-011 — Concorrência resolvida no PostgreSQL

**Status:** aceita.

**Decisão:** baixa usa atualização atômica condicionada a saldo suficiente, na
mesma transação das movimentações e idempotência.

**Consequências:** duas notas concorrentes não tornam saldo negativo. A regra
permanece no domínio e a garantia de corrida fica na persistência.

## ADR-012 — Retry fixo, limitado e DLQ

**Status:** aceita.

**Decisão:** falhas temporárias passam por fila de retry com atraso fixo e limite
configurável. Mensagens inválidas, terminais ou esgotadas seguem para DLQ.

**Consequências:** evita loop quente e mantém a solução compreensível. Backoff
exponencial e múltiplas filas foram evitados por excederem o escopo atual.

## ADR-013 — Produto inativo em vez de exclusão

**Status:** aceita.

**Decisão:** produtos são ativados ou inativados; não existe endpoint de delete.

**Consequências:** notas e movimentações preservam rastreabilidade. Consultas e
criação de notas precisam considerar o estado ativo.

## ADR-014 — Snapshot do produto na nota

**Status:** aceita.

**Decisão:** item guarda ID, código, descrição e valor unitário obtidos do
Estoque durante criação ou edição.

**Consequências:** histórico não muda junto ao cadastro. Existe duplicação
intencional e não há foreign key cruzando schemas.

## ADR-015 — Valores monetários em ponto flutuante

**Status:** aceita por requisito de implementação.

**Decisão:** produto, item e totais usam valor decimal representado em ponto
flutuante, sem o antigo sufixo `em_centavos`.

**Consequências:** API mais direta para o exercício. Em um sistema fiscal real,
`NUMERIC`/decimal ou inteiros em unidade mínima seriam preferíveis para evitar
imprecisão binária.

## ADR-016 — Swagger versionado

**Status:** aceita.

**Decisão:** anotações Gin geram OpenAPI e Swagger UI em cada serviço; artefatos
gerados ficam versionados.

**Consequências:** contrato pode ser consultado sem instalar o gerador. Mudanças
de handlers exigem regeneração explícita.

## ADR-017 — Docker Compose completo

**Status:** aceita.

**Decisão:** imagens multi-stage e sem root, migrations separadas, health checks
e dependências explícitas.

**Consequências:** o projeto sobe com poucos comandos e reproduz a topologia. A
primeira compilação consome mais tempo e recursos que a execução local direta.

## ADR-018 — Angular 22 standalone

**Status:** aceita.

**Decisão:** usar a versão estável ativa mais recente do Angular na criação do
projeto, com componentes standalone, modo estrito e rotas lazy por feature.

**Consequências:** estrutura alinhada às APIs atuais e menor dependência de
NgModules. Angular 22 exige versões recentes de Node e TypeScript; o projeto
declara essas versões no `package.json` e no `package-lock.json`. npm é o
gerenciador oficial do workspace para reduzir pré-requisitos e acompanhar a
distribuição padrão do Node.js.

## ADR-019 — RxJS como padrão de estado do frontend

**Status:** aceita.

**Decisão:** stores por feature mantêm um `BehaviorSubject` privado, expõem
somente `Observable` e alteram estado por métodos explícitos. O template usa
`AsyncPipe` para controlar subscriptions.

**Consequências:** estado reativo sem adicionar NgRx no início do projeto. A
equipe precisa evitar mutação direta, Subjects públicos e subscriptions manuais
sem ciclo de vida controlado. Se a complexidade crescer, a decisão pode ser
reavaliada.

## ADR-020 — Angular Material para componentes visuais

**Status:** aceita.

**Decisão:** usar Angular Material e CDK na mesma linha principal do Angular.
Componentes são importados por módulos específicos nos componentes standalone.

**Consequências:** UI consistente, acessível e integrada ao ecossistema oficial
do Angular, sem depender do licenciamento de outra biblioteca visual. O bundle
é controlado por imports locais e customizações devem usar o sistema de temas
antes de sobrescritas frágeis com `::ng-deep`.

## ADR-021 — Proxy local do Angular para as APIs

**Status:** substituída pela ADR-023.

**Decisão:** o frontend usa os prefixos relativos `/api/estoque` e
`/api/faturamento`. Durante o desenvolvimento, `proxy.conf.json` encaminha as
requisições para as portas locais dos serviços.

**Consequências:** o navegador não enfrenta bloqueio de CORS no desenvolvimento
e o código não contém hosts absolutos. O servidor ou proxy reverso de produção
deve publicar os mesmos prefixos para que o frontend não precise ser recompilado
por ambiente.

## ADR-022 — Componentes frontend por responsabilidade

**Status:** aceita.

**Decisão:** features do frontend separam componentes visuais em pastas como
`form/`, `list/` e `details/`. Elementos independentes do domínio, incluindo
paginação e feedback de dados, ficam em `shared/ui` e usam inputs e outputs.

**Consequências:** páginas atuam como containers de orquestração, componentes de
feature ficam menores e a mesma paginação pode ser aplicada a Produtos e Notas
Fiscais. Um componente só deve ir para `shared` quando não importar models ou
stores de uma feature específica.

## ADR-023 — CORS explícito nas APIs

**Status:** aceita.

**Decisão:** o frontend chama diretamente as URLs de Estoque e Faturamento. Cada
API recebe uma lista obrigatória de origens permitidas, separadas por vírgula,
pelas variáveis `ESTOQUE_CORS_ALLOWED_ORIGINS` e
`FATURAMENTO_CORS_ALLOWED_ORIGINS`. O proxy local do Angular foi removido.

**Consequências:** a política de acesso pertence às APIs e funciona para qualquer
cliente web autorizado, não apenas para `ng serve`. Novas origens precisam ser
declaradas explicitamente no ambiente; preflights de origens desconhecidas
recebem HTTP 403.

## ADR-024 — Sistema visual compacto e confirmações Material

**Status:** aceita.

**Decisão:** componentes usam uma identidade visual global baseada em variáveis
de cor, superfícies, sombras e densidade compacta. Paginações do Angular Material
são traduzidas globalmente para português. Perguntas de confirmação usam o
componente compartilhado `ConfirmationDialog`, nunca caixas nativas do browser.

**Consequências:** telas exibem mais informação sem perder hierarquia visual e
novas features herdam paginação e diálogos consistentes. Alterações de identidade
visual devem começar nas variáveis globais; estilos locais ficam reservados às
necessidades específicas de cada componente.

## ADR-025 — Polling não bloqueante do fechamento de notas

**Status:** aceita.

**Decisão:** após iniciar a impressão/fechamento, o frontend acompanha cada nota
`PROCESSANDO` com um fluxo RxJS independente a cada 1,5 segundo. IDs são
deduplicados e `exhaustMap` evita sobreposição de requisições. O fluxo termina no
estado final ou quando o componente é destruído.

**Consequências:** a linha muda para `FECHADA` ou volta para `ABERTA` com o motivo
da rejeição sem recarregar a página e sem bloquear outras interações. O custo de
rede fica restrito às notas realmente em processamento. WebSocket ou SSE podem
substituir o polling se a escala futura justificar infraestrutura adicional.

## ADR-026 — Cores semânticas nas ações das tabelas

**Status:** aceita.

**Decisão:** as tabelas de Produtos e Notas Fiscais usam botões compactos com
cores consistentes por intenção: azul para consulta, âmbar para edição, verde ou
vermelho para alteração de situação e roxo para impressão. A cor complementa,
mas não substitui, ícones, tooltips e rótulos acessíveis. A tabela usa linhas
alternadas e realce no hover para facilitar a leitura horizontal.

**Consequências:** ações distintas são reconhecidas mais rapidamente sem perder
acessibilidade. Botões indisponíveis ficam neutros e não sugerem que podem ser
acionados. Novas tabelas devem reutilizar a mesma semântica visual. Quantidades
e valores recebem realces compactos e as datas de cadastro e atualização são
apresentadas separadamente para preservar a leitura temporal dos registros.

## ADR-027 — Seleção de produtos ativos na nota fiscal

**Status:** aceita.

**Decisão:** o formulário de nota fiscal consulta a API de Estoque e oferece um
select contendo somente produtos ativos. A consulta percorre todas as páginas da
API, respeitando o limite de 100 registros por requisição. As opções mostram
código, descrição e saldo, enquanto o comando enviado ao Faturamento preserva
somente código e quantidade.

**Consequências:** o usuário não precisa digitar ou memorizar códigos e consegue
avaliar o saldo antes de incluir o item. A API de Faturamento continua responsável
pela validação definitiva do produto e do estoque, evitando usar o saldo exibido
no navegador como regra de negócio. A pesquisa por código ou descrição ocorre
localmente sobre os produtos já carregados, evitando tráfego a cada tecla. Em
edições, itens cujo produto deixou de estar ativo preservam seu snapshot para
identificação visual, mas precisam ser substituídos ou removidos antes de salvar.

## ADR-028 — Ordenação operacional das listagens

**Status:** aceita.

**Decisão:** a API de Estoque lista produtos pela data de cadastro decrescente.
A API de Faturamento prioriza notas `PROCESSANDO`, depois `ABERTA` e por último
`FECHADA`. Datas de cadastro e identificadores estáveis são usados como critérios
secundários.

**Consequências:** registros recentes ficam visíveis primeiro no estoque e notas
que demandam acompanhamento aparecem antes das demais. Os desempates explícitos
mantêm a paginação determinística mesmo quando vários registros compartilham a
mesma data ou status.

## Decisões adiadas

- autenticação e autorização;
- API Gateway;
- métricas, tracing e agregação de logs;
- múltiplas instâncias do worker de Outbox;
- bancos físicos separados;
- geração de documento fiscal ou integração governamental.

Itens adiados não devem ser lidos como parte implementada da solução atual.
