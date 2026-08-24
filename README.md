# Sistema de Emissão de Notas Fiscais

Monorepo de estudo com dois microsserviços Go para cadastro de produtos,
controle de estoque e emissão assíncrona de notas fiscais.

## Visão geral

O sistema separa dois contextos de negócio:

- **Estoque:** produtos, saldo, movimentações e baixa concorrente;
- **Faturamento:** notas fiscais, itens, totalizadores e fechamento.

As APIs usam Go, Gin, GORM e PostgreSQL. O fechamento de uma nota é processado
de forma assíncrona pelo RabbitMQ, com Transactional Outbox, idempotência,
retentativas limitadas e DLQ.

```mermaid
flowchart LR
    C[Cliente] -->|HTTP| E[Estoque]
    C -->|HTTP| F[Faturamento]
    F -->|solicita baixa| R[RabbitMQ]
    R --> E
    E -->|resultado| R
    R --> F
    E --> P[(PostgreSQL)]
    F --> P
```

## Estado do projeto

| Componente | Estado |
|---|---|
| Estoque Service | Implementado |
| Faturamento Service | Implementado |
| Fluxo assíncrono | Implementado |
| Testes unitários, integrados e E2E | Implementados |
| Docker Compose completo | Implementado |
| Frontend Angular | Pendente |

## Início rápido com Docker

Pré-requisito: Docker Desktop com Docker Compose.

Crie o arquivo de ambiente.

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

Suba a aplicação completa a partir da raiz:

```console
docker compose config
docker compose up -d --build
docker compose ps
```

Na primeira execução, o Docker compila as duas APIs. Os containers de migration
aplicam os arquivos SQL antes da inicialização dos serviços.

## Endereços locais

| Recurso | Endereço |
|---|---|
| Estoque | <http://localhost:8081> |
| Swagger do Estoque | <http://localhost:8081/swagger/index.html> |
| Faturamento | <http://localhost:8082> |
| Swagger do Faturamento | <http://localhost:8082/swagger/index.html> |
| RabbitMQ Management | <http://localhost:15672> |

Credenciais e portas de desenvolvimento estão declaradas no `.env.example`.

## Verificação rápida

```console
docker compose logs -f estoque faturamento
```

Com a pilha saudável, execute os testes de ponta a ponta:

```console
cd tests/e2e
go test -count=1 -v ./...
```

## Estrutura

```text
.
├── docs/                   documentação técnica
├── infrastructure/         inicialização do PostgreSQL
├── services/
│   ├── estoque/            módulo Go do Estoque
│   └── faturamento/        módulo Go do Faturamento
├── tests/e2e/              testes externos do fluxo completo
├── .env.example            contrato de configuração
└── docker-compose.yml      ambiente local completo
```

## Documentação

O índice completo está em [docs/README.md](docs/README.md).

- [Arquitetura da solução](docs/README_ARQUITETURA.md)
- [Decisões arquiteturais](docs/README_DECISOES_ARQUITETURAIS.md)
- [Modelo de domínio](docs/README_MODELO_DOMINIO.md)
- [Serviço de Estoque](docs/SERVICO_ESTOQUE.md)
- [Serviço de Faturamento](docs/SERVICO_FATURAMENTO.md)
- [Mensageria e resiliência](docs/RESILIENCIA.md)
- [Execução e Docker](docs/EXECUCAO_DOCKER.md)
- [Testes](docs/TESTES.md)

## Comandos mais usados

```console
docker compose up -d --build
docker compose ps
docker compose logs -f estoque faturamento
docker compose down
```

Os comandos locais, migrations, exemplos HTTP e procedimentos de diagnóstico
ficam nos documentos específicos para evitar duplicação neste README.
