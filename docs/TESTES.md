# Estratégia de testes

## Visão geral

O projeto possui três níveis de teste:

| Nível | Objetivo | Dependências externas |
|---|---|---|
| Unitário | Regras de domínio, use cases e auxiliares | Nenhuma |
| Integração | Repositórios, transações e PostgreSQL real | PostgreSQL |
| End-to-end | Fluxo externo completo | Duas APIs, PostgreSQL e RabbitMQ |

## Testes unitários

Cada microsserviço é um módulo Go independente.

Estoque:

```console
cd services/estoque
go test ./...
go vet ./...
```

Faturamento:

```console
cd services/faturamento
go test ./...
go vet ./...
```

Esses testes cobrem, entre outros pontos:

- criação e alteração das entidades;
- invariantes de saldo, valor e quantidade;
- estados da nota fiscal;
- normalização de texto;
- paginação e filtros;
- orquestração dos use cases;
- tradução de erros HTTP;
- leitura dos headers de retry.

### Frontend Angular

PowerShell, Bash, Zsh, Git Bash ou Prompt de Comando:

```console
cd frontend/web
npm test -- --watch=false
```

A suíte usa Angular TestBed com Vitest e cobre clientes HTTP, stores RxJS,
filtros, tabelas, formulários, diálogos, paginação e estados de feedback. Polling
e temporizadores são controlados com relógio falso para não tornar os testes
lentos ou dependentes da rede.

## Testes de integração

Requerem PostgreSQL ativo e migrations aplicadas. A build tag impede que sejam
executados acidentalmente junto aos testes unitários.

Estoque:

```console
cd services/estoque
go test -tags=integration ./internal/infrastructure/repository -count=1 -v
```

Faturamento:

```console
cd services/faturamento
go test -tags=integration ./internal/infrastructure/repository -count=1 -v
```

Os testes verificam persistência real, reconstituição de domínio, transações,
movimentações, Outbox e idempotência.

## Testes end-to-end

Localização: `tests/e2e`.

Suba a pilha completa:

```console
docker compose up -d --build
docker compose ps
```

PowerShell:

```powershell
Set-Location tests/e2e
go test -count=1 -v ./...
```

Bash, Zsh, Git Bash ou Prompt de Comando:

```console
cd tests/e2e
go test -count=1 -v ./...
```

### Configuração E2E

```text
E2E_ESTOQUE_URL
E2E_FATURAMENTO_URL
E2E_RABBITMQ_URL
E2E_DATABASE_URL
E2E_TIMEOUT
```

As variáveis ficam no `.env` da raiz e não possuem fallback no código de teste.

### Cenários cobertos

1. saldo suficiente fecha a nota, reduz o saldo e registra saída;
2. saldo insuficiente reabre a nota, persiste o motivo e preserva o saldo;
3. produto inativado depois da criação provoca rejeição;
4. falha em um item reverte a baixa de todos os itens;
5. solicitação duplicada não reduz saldo duas vezes;
6. resultado duplicado não repete a transição da nota;
7. mensagens inválidas chegam às DLQs de Estoque e Faturamento.

Os dados recebem códigos únicos para permitir repetição da suíte no mesmo banco.

### E2E visual do frontend

Localização: `frontend/web/e2e`.

Com a infraestrutura e as APIs saudáveis:

```console
cd frontend/web
npx playwright install chromium
npm run test:e2e
```

A suíte usa seletores acessíveis por papel e rótulo, execução serial e dados
únicos. Ela valida cadastro de produto, criação da nota pelo seletor pesquisável,
fechamento assíncrono até `FECHADA` e bloqueio visual de itens inativos. O servidor
Angular é iniciado automaticamente pelo Playwright.

Em falhas são preservados screenshot, vídeo e trace. Use `npm run test:e2e:ui`
para depuração interativa.

## Por que o teste de indisponibilidade não derruba o RabbitMQ

A reconexão automática existe, mas não há um teste que interrompe o container.
Esse cenário é disruptivo: interfere em outras suítes e em qualquer pessoa que
esteja usando o ambiente. Para o escopo do projeto, a resiliência é validada por
testes dos mecanismos isolados e o fluxo completo é validado sem derrubar a
infraestrutura compartilhada.

## Ordem recomendada antes de entregar

```text
1. npm test -- --watch=false no frontend
2. go test ./... nos dois módulos
3. go vet ./... nos dois módulos
4. testes de integração com PostgreSQL
5. docker compose config
6. docker compose up -d --build
7. testes E2E das APIs
8. npm run test:e2e no frontend
9. git diff --check
```

## Diagnóstico de falhas

- `connection refused` em 8081/8082: APIs não estão ativas ou saudáveis;
- falha em 5432: PostgreSQL não está pronto ou a URL está incorreta;
- falha em 5672: RabbitMQ não está pronto ou as credenciais estão incorretas;
- timeout esperando status: consulte logs dos dois workers e filas no Management;
- migration ausente: execute os containers de migration antes das APIs.
