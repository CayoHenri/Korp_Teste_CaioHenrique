# Execução e Docker

## Pré-requisitos

- Docker Desktop com Docker Compose;
- portas `5432`, `5672`, `15672`, `8081` e `8082` livres;
- arquivo `.env` criado a partir de `.env.example`.

PowerShell:

```powershell
Copy-Item .env.example .env
```

Bash, Zsh ou Git Bash:

```bash
cp .env.example .env
```

Prompt de Comando:

```bat
copy .env.example .env
```

O `.env.example` é o contrato versionado. O `.env` contém a configuração local
e não deve ser commitado.

## Serviços do Compose

| Serviço | Tipo | Função |
|---|---|---|
| `postgres` | contínuo | Banco dos dois schemas |
| `rabbitmq` | contínuo | Broker e interface administrativa |
| `estoque-migrations` | execução única | Atualizar schema `estoque` |
| `faturamento-migrations` | execução única | Atualizar schema `faturamento` |
| `estoque` | contínuo | API e worker de baixas |
| `faturamento` | contínuo | API, Outbox e worker de resultados |

## Subir a aplicação completa

Os comandos são iguais em PowerShell, Bash, Zsh, Git Bash e Prompt de Comando.
Execute na raiz:

```console
docker compose config
docker compose up -d --build
docker compose ps
```

`config` valida variáveis e sintaxe. `--build` recompila imagens quando houver
mudança. Na primeira execução, downloads e compilação tornam o processo mais
demorado.

## Ordem de inicialização

O Compose usa condições de dependência:

```text
PostgreSQL saudável ──> migrations concluídas ──> APIs
RabbitMQ saudável ──────────────────────────────> APIs
Estoque saudável ───────────────────────────────> Faturamento
```

As migrations são processos separados. A API não altera schema durante o
startup.

## Imagens

Cada serviço usa um Dockerfile multi-stage:

1. `golang:1.25-alpine` baixa módulos e compila `api` e `migrate`;
2. `alpine` recebe apenas binários, migrations, certificados e timezone;
3. a execução final usa o usuário `app`, sem privilégios de root.

A mesma imagem é utilizada pela API e pelo container de migration. O Compose
altera apenas o comando do container de migration.

## Endereços

| Recurso | Host local | Rede Docker |
|---|---|---|
| PostgreSQL | `localhost:5432` | `postgres:5432` |
| RabbitMQ | `localhost:5672` | `rabbitmq:5672` |
| Estoque | `localhost:8081` | `estoque:8081` |
| Faturamento | `localhost:8082` | `faturamento:8082` |
| RabbitMQ Management | `localhost:15672` | `rabbitmq:15672` |

Dentro de um container, `localhost` aponta para o próprio container. Por isso,
as URLs internas usam os nomes dos serviços.

## Operação diária

Logs das APIs:

```console
docker compose logs -f estoque faturamento
```

Logs de migrations:

```console
docker compose logs estoque-migrations faturamento-migrations
```

Reconstruir um serviço:

```console
docker compose up -d --build estoque
docker compose up -d --build faturamento
```

Executar migrations novamente:

```console
docker compose run --rm estoque-migrations
docker compose run --rm faturamento-migrations
```

Parar preservando dados:

```console
docker compose down
```

## Limpeza de dados

Este comando remove containers e volumes, apagando PostgreSQL e RabbitMQ locais:

```console
docker compose down --volumes
```

Use apenas quando quiser reconstruir o ambiente do zero. O script em
`infrastructure/postgres/init` só é executado na criação de um volume vazio e
cria os schemas; as tabelas continuam sendo responsabilidade das migrations.

## Desenvolvimento híbrido

Para executar as APIs com `go run` e manter apenas a infraestrutura no Docker:

```console
docker compose up -d postgres rabbitmq
```

Depois, em terminais separados:

```console
cd services/estoque
go run ./cmd/migrate up
go run ./cmd/api
```

```console
cd services/faturamento
go run ./cmd/migrate up
go run ./cmd/api
```

As URLs locais do `.env.example` usam `localhost`, enquanto o Compose injeta
URLs internas próprias nos containers.

## Diagnóstico

### Porta ocupada

Se o Docker informar `ports are not available`, encerre a aplicação local que
usa a mesma porta ou altere a configuração antes de subir o Compose.

PowerShell para identificar o processo:

```powershell
Get-NetTCPConnection -LocalPort 8081,8082 -State Listen |
    Select-Object LocalPort, OwningProcess
```

### Migration falhou

```console
docker compose logs estoque-migrations
docker compose logs faturamento-migrations
```

Consulte o documento do serviço antes de usar `force`. O comando altera a
versão registrada, mas não corrige o schema.

### API não fica saudável

```console
docker compose ps
docker compose logs estoque
docker compose logs faturamento
```

Os endpoints `/health` verificam a conexão com PostgreSQL e retornam `503` se o
banco estiver indisponível.
