# Sistema de Emissão de Notas Fiscais

Sistema distribuído de estudo para cadastro de produtos, controle de estoque e
emissão de notas fiscais. A solução é organizada como monorepo e utiliza Angular,
microsserviços Go, PostgreSQL e RabbitMQ.

## Estado atual

| Componente | Estado |
|---|---|
| PostgreSQL e RabbitMQ | Implementados no Docker Compose |
| Estoque Service | Cadastro, listagem e consulta de produtos implementados |
| Faturamento Service | Planejado, ainda não implementado |
| Frontend Angular | Planejado, ainda não implementado |

## Decisões gerais

- Cada microsserviço possui seu próprio módulo Go e poderá ser implantado de forma independente.
- A instância PostgreSQL é compartilhada, mas cada serviço é proprietário de seu schema: `estoque` ou `faturamento`.
- Alterações de schema são explícitas e versionadas com `golang-migrate`.
- GORM é usado para persistência, sem `AutoMigrate`.
- As APIs HTTP usam Gin e são documentadas com Swagger/OpenAPI.
- Não existem valores de configuração padrão no código ou no Compose. Variáveis obrigatórias ausentes causam erro imediatamente.
- RabbitMQ será usado na integração assíncrona entre Faturamento e Estoque.

## Pré-requisitos

- Docker Desktop com Docker Compose;
- Go 1.25 ou compatível com a versão declarada pelos módulos;
- portas `5432`, `5672`, `15672` e `8081` livres para a configuração de exemplo.

## Configuração do ambiente

Crie o arquivo local `.env` a partir do exemplo:

PowerShell:

```powershell
Copy-Item .env.example .env
```

Bash, Zsh ou Git Bash:

```bash
cp .env.example .env
```

Prompt de Comando do Windows (`cmd.exe`):

```bat
copy .env.example .env
```

O `.env` é ignorado pelo Git e pode conter valores locais. O `.env.example` é
versionado e funciona como contrato das variáveis necessárias, sem armazenar
segredos reais de outros ambientes.

O Docker Compose lê o `.env` automaticamente quando executado na raiz. Os
comandos Go do Estoque também carregam o `.env` da raiz quando executados dentro
de `services/estoque`. Variáveis já definidas no processo têm precedência sobre
o arquivo.

## Infraestrutura: PostgreSQL e RabbitMQ

### Decisões

- PostgreSQL mantém os dados em volume nomeado.
- RabbitMQ mantém filas e metadados em volume nomeado.
- Os containers compartilham a rede `korp-network`.
- Health checks verificam se os serviços estão realmente prontos.
- O script inicial cria somente os schemas. As tabelas pertencem às migrations de cada microsserviço.

### Iniciar e verificar

Os comandos abaixo são iguais no PowerShell, Bash, Zsh, Git Bash e Prompt de
Comando. Execute-os na raiz do repositório:

```console
docker compose config
docker compose up -d
docker compose ps
```

Para acompanhar logs:

```console
docker compose logs -f postgres rabbitmq
```

Serviços da configuração de exemplo:

| Serviço | Endereço | Finalidade |
|---|---|---|
| PostgreSQL | `localhost:5432` | Banco relacional |
| RabbitMQ | `localhost:5672` | Protocolo AMQP usado pelas aplicações |
| RabbitMQ Management | <http://localhost:15672> | Interface administrativa |

Dentro da rede Docker, aplicações usam `postgres:5432` e `rabbitmq:5672`, não
`localhost`, porque cada container possui seu próprio endereço local.

### Parar ou reinicializar

Parar preservando os dados:

```console
docker compose down
```

Remover containers e volumes, apagando todos os dados locais:

```console
docker compose down --volumes
```

O script `infrastructure/postgres/init/001-create-schemas.sql` é executado
somente quando o volume do PostgreSQL é criado vazio. Mudanças posteriores de
tabelas devem ser feitas por migrations, não por esse script.

## Microsserviço de Estoque

Localização: `services/estoque`.

### Responsabilidade

O Estoque Service será o único responsável por produtos, saldos, movimentações,
baixa concorrente de estoque e idempotência do consumo de mensagens.

Nesta etapa estão implementados:

- conexão PostgreSQL por GORM;
- endpoint `GET /health`;
- documentação Swagger;
- configuração obrigatória pela infraestrutura;
- comando independente de migrations;
- migration inicial de `estoque.produtos`;
- domínio e casos de uso de Produto;
- repositório de produtos com GORM;
- cadastro, listagem e consulta por ID ou código;
- encerramento gracioso da API.

### Decisões técnicas

- Gin expõe a camada HTTP, sem entrar no domínio.
- GORM será usado por repositórios e transações.
- `AutoMigrate` não é usado, evitando mudanças implícitas no banco.
- `golang-migrate` aplica arquivos SQL `up` e `down` em ordem de versão.
- O health check executa `PingContext` no banco e retorna `503` quando a conexão não está disponível.
- Swagger é gerado a partir de anotações e seus artefatos são versionados.

### Variáveis obrigatórias

```text
ESTOQUE_HTTP_PORT
ESTOQUE_DATABASE_URL
```

Opcionalmente, as variáveis podem ser sobrescritas apenas na sessão atual.

PowerShell:

```powershell
$env:ESTOQUE_HTTP_PORT = "8081"
$env:ESTOQUE_DATABASE_URL = "postgres://korp:korp_dev_password@localhost:5432/korp_db?sslmode=disable"
```

Bash, Zsh ou Git Bash:

```bash
export ESTOQUE_HTTP_PORT="8081"
export ESTOQUE_DATABASE_URL="postgres://korp:korp_dev_password@localhost:5432/korp_db?sslmode=disable"
```

Prompt de Comando do Windows (`cmd.exe`):

```bat
set ESTOQUE_HTTP_PORT=8081
set ESTOQUE_DATABASE_URL=postgres://korp:korp_dev_password@localhost:5432/korp_db?sslmode=disable
```

Normalmente não é necessário executar esses comandos, pois o serviço carrega o
`.env` da raiz. Não versione credenciais privadas.

### Dependências, testes e análise estática

Os comandos são iguais nos terminais suportados:

```console
cd services/estoque
go mod download
go test ./...
go vet ./...
```

### Comandos de migration

Os comandos devem ser executados dentro de `services/estoque`, pois o caminho
`file://migrations` é resolvido a partir do diretório atual.

Aplicar todas as migrations pendentes:

```console
go run ./cmd/migrate up
```

Reverter exatamente a última migration aplicada:

```console
go run ./cmd/migrate down
```

Consultar versão e estado `dirty`:

```console
go run ./cmd/migrate version
```

Corrigir manualmente a versão registrada:

```console
go run ./cmd/migrate force 1
```

`force` não executa SQL. Ele altera apenas a versão registrada e deve ser usado
somente para recuperar uma migration interrompida, depois de conferir o estado
real do banco.

Fluxo normal de desenvolvimento:

```console
go run ./cmd/migrate up
go run ./cmd/api
```

Para criar uma nova migration, adicione o par com o próximo número sequencial:

```text
migrations/000002_descricao_da_alteracao.up.sql
migrations/000002_descricao_da_alteracao.down.sql
```

O arquivo `up` aplica a mudança. O arquivo `down` deve desfazer apenas essa
mudança de maneira coerente.

### API e Swagger

Com o serviço executando:

| Recurso | Endereço |
|---|---|
| Health check | <http://localhost:8081/health> |
| Swagger UI | <http://localhost:8081/swagger/index.html> |

Endpoints implementados:

| Método | Rota | Resultado |
|---|---|---|
| `POST` | `/produtos` | Cadastra um produto |
| `GET` | `/produtos` | Lista produtos ordenados por código |
| `GET` | `/produtos/{id}` | Consulta por UUID |
| `GET` | `/produtos/codigo/{codigo}` | Consulta por código |

Exemplo de cadastro, válido em Bash, Zsh e Git Bash:

```bash
curl -X POST http://localhost:8081/produtos \
  -H "Content-Type: application/json" \
  -d '{"codigo":"SKU-001","descricao":"Teclado mecanico","saldo":10}'
```

PowerShell:

```powershell
$body = @{
    codigo = "SKU-001"
    descricao = "Teclado mecanico"
    saldo = 10
} | ConvertTo-Json

Invoke-RestMethod `
    -Method Post `
    -Uri "http://localhost:8081/produtos" `
    -ContentType "application/json" `
    -Body $body
```

Depois de alterar anotações HTTP, regenere os arquivos Swagger:

```console
go run github.com/swaggo/swag/cmd/swag@v1.16.6 init -g cmd/api/main.go -o docs
```

Os arquivos gerados são `docs/docs.go`, `docs/swagger.json` e
`docs/swagger.yaml` dentro do módulo de Estoque.

## Microsserviço de Faturamento

Localização planejada: `services/faturamento`.

### Responsabilidade

Será responsável por notas fiscais, itens, numeração sequencial, estados
`ABERTA`, `PROCESSANDO` e `FECHADA`, Transactional Outbox e processamento dos
resultados publicados pelo Estoque.

### Decisões previstas

- módulo Go independente;
- Gin, GORM, Swagger e `golang-migrate`, seguindo o padrão do Estoque;
- propriedade exclusiva do schema `faturamento`;
- publicação de solicitações de baixa pelo RabbitMQ;
- confirmação ou reabertura da nota a partir do resultado do Estoque.

O serviço ainda não foi criado. Portanto, ainda não existem comandos de build,
migration ou execução para ele. Esta seção será atualizada junto da implementação.

## Frontend Angular

Localização planejada: `frontend/web`.

Será responsável pelo cadastro e consulta de produtos, criação e consulta de
notas, solicitação de fechamento e feedback de processamento assíncrono. O
frontend ainda não foi criado e não possui comandos executáveis nesta etapa.

## Estrutura atual

```text
Korp_Teste_CaioHenrique/
├── docs/
├── infrastructure/
│   └── postgres/init/
├── services/
│   └── estoque/
│       ├── cmd/api/
│       ├── cmd/migrate/
│       ├── docs/
│       ├── internal/infrastructure/
│       ├── internal/presentation/http/
│       └── migrations/
├── .env.example
├── docker-compose.yml
└── README.md
```

## Documentação arquitetural

- [Arquitetura](docs/README_ARQUITETURA.md)
- [Decisões arquiteturais](docs/README_DECISOES_ARQUITETURAIS.md)
- [Modelo de domínio](docs/README_MODELO_DOMINIO.md)
