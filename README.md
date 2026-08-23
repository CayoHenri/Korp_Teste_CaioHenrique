# Sistema de Emissão de Notas Fiscais

Sistema distribuído de estudo para cadastro de produtos, controle de estoque e
emissão de notas fiscais. A solução é organizada como monorepo e utiliza Angular,
microsserviços Go, PostgreSQL e RabbitMQ.

## Estado atual

| Componente | Estado |
|---|---|
| PostgreSQL e RabbitMQ | Implementados no Docker Compose |
| Estoque Service | Produtos, movimentações e baixa transacional implementados |
| Faturamento Service | Notas, itens e início de fechamento com Outbox implementados |
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
- models GORM separados dos repositórios e conversão explícita para domínio;
- DTOs HTTP separados dos handlers;
- normalização compartilhada de textos com remoção de espaços e uppercase;
- injeção de dependências centralizada em um container próprio;
- envelopes HTTP e tradução de erros de domínio centralizados;
- entidades com estado privado, leitura por getters e alterações por métodos de negócio;
- camada application organizada em um use case por operação;
- cadastro, listagem e consulta por ID ou código;
- ativação e inativação sem exclusão física;
- movimentações auditáveis de entrada e saída;
- baixa transacional, concorrente e idempotente preparada para mensageria;
- encerramento gracioso da API.

### Decisões técnicas

- Gin expõe a camada HTTP, sem entrar no domínio.
- GORM será usado por repositórios e transações.
- `AutoMigrate` não é usado, evitando mudanças implícitas no banco.
- `golang-migrate` aplica arquivos SQL `up` e `down` em ordem de versão.
- O health check executa `PingContext` no banco e retorna `503` quando a conexão não está disponível.
- Swagger é gerado a partir de anotações e seus artefatos são versionados.
- Construtores de entidades seguem o padrão `NewNomeDaEntidade`.
- Models GORM ficam em `internal/infrastructure/database/models` e expõem `ToDomain`.
- DTOs ficam na apresentação HTTP, separados dos handlers.
- Código e descrição de produtos são persistidos em uppercase.
- `Produto` não expõe campos públicos; getters seguem a convenção idiomática de Go, como `ID()` e `Descricao()`.
- Não são criados setters genéricos; o caso de uso de atualização orquestra os métodos de domínio `AtualizarDescricao` e `AtualizarSaldo`.
- Dados persistidos são reconstituídos por `NewProdutoWithState`, que reaplica as invariantes.
- Produtos novos iniciam ativos e não possuem operação de exclusão; o ciclo de vida usa `Ativar` e `Inativar`.
- Alterações manuais de saldo geram movimentações de ajuste do tipo `ENTRADA` ou `SAIDA`.
- Valores monetários são armazenados como ponto flutuante; o Produto mantém `valor`.
- Baixas de nota fiscal usam atualização atômica, transação única e idempotência por `eventId`.
- Repositories usam nomes de negócio, como `ProdutoRepository`, sem prefixo da tecnologia de persistência.
- O pacote `internal/dependency` concentra a composição de repositories, use cases, handlers e router.
- A camada application não usa services genéricos: cada operação possui seu próprio use case com método `Execute`.
- O pacote HTTP `response` fornece respostas `OK`, `Created`, `Data`, `Message` e `Error`.
- O pacote HTTP `domainerror` traduz erros conhecidos do domínio e oculta detalhes de falhas inesperadas.

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

Com PostgreSQL ativo e migrations aplicadas, execute também os testes integrados:

```console
go test -tags=integration ./internal/infrastructure/repository -count=1 -v
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
migrations/000005_descricao_da_alteracao.up.sql
migrations/000005_descricao_da_alteracao.down.sql
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
| `PUT` | `/produtos/{id}` | Atualiza descrição e saldo |
| `PATCH` | `/produtos/{id}/ativar` | Ativa um produto |
| `PATCH` | `/produtos/{id}/inativar` | Inativa sem excluir |
| `GET` | `/produtos/{id}/movimentacoes` | Lista o histórico de estoque |

A listagem de produtos aceita `pagina`, `tamanhoPagina` e os filtros opcionais
`codigo`, `descricao` e `ativo`:

```text
GET /produtos?pagina=1&tamanhoPagina=20&descricao=teclado&ativo=true
```

Quando omitidos, são usados `pagina=1` e `tamanhoPagina=20`. O tamanho máximo
permitido por página é 100.

Não existe endpoint `DELETE /produtos/{id}`. Produtos referenciados por
movimentações ou notas precisam continuar disponíveis para rastreabilidade.

Respostas de sucesso seguem o envelope:

```json
{
  "success": true,
  "data": {}
}
```

Respostas de erro seguem o envelope:

```json
{
  "success": false,
  "error": {
    "code": "PRODUTO_NAO_ENCONTRADO",
    "message": "produto nao encontrado"
  }
}
```

Erros inesperados retornam `ERRO_INTERNO` sem expor mensagens de banco, stack
traces ou outros detalhes internos.

Exemplo de cadastro, válido em Bash, Zsh e Git Bash:

```bash
curl -X POST http://localhost:8081/produtos \
  -H "Content-Type: application/json" \
  -d '{"codigo":"SKU-001","descricao":"Teclado mecanico","saldo":10,"valor":159.90}'
```

PowerShell:

```powershell
$body = @{
    codigo = "SKU-001"
    descricao = "Teclado mecanico"
    saldo = 10
    valor = 159.90
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

Localização: `services/faturamento`.

### Responsabilidade

É responsável por notas fiscais, itens, numeração sequencial e pelo ciclo de
fechamento. Nesta etapa, o serviço cria, lista e consulta notas e inicia o
fechamento com uma Transactional Outbox.

### Implementado

- módulo Go independente;
- Gin, GORM, Swagger e `golang-migrate`, seguindo o padrão do Estoque;
- propriedade exclusiva do schema `faturamento`;
- status da nota persistido no enum PostgreSQL `faturamento.nota_fiscal_status`;
- domínio encapsulado para `NotaFiscal` e `ItemNotaFiscal`;
- um use case por operação;
- numeração sequencial por sequence do PostgreSQL;
- controle de migrations isolado em `faturamento_schema_migrations`;
- criação de notas no estado `ABERTA` com snapshot dos produtos;
- validação síncrona do produto pelo código na API do Estoque;
- inclusão permitida somente para produtos ativos;
- snapshot do ID, código, descrição e valor unitário em cada item;
- nome e endereço do cliente no cabeçalho;
- normalização compartilhada de código, descrição, nome e endereço com trim e uppercase;
- totalizadores de quantidade e valor calculados pelo domínio;
- listagem e consulta de notas;
- transição `ABERTA` para `PROCESSANDO`;
- criação do evento Outbox na mesma transação da mudança de estado;
- worker periódico para publicação dos eventos pendentes da Outbox;
- publicação persistente de `estoque.baixa.solicitada` no RabbitMQ;
- confirmação do broker antes de marcar `published_at`, com retry na próxima execução;
- DTOs, respostas HTTP e tradução de erros separados;
- injeção de dependências centralizada e encerramento gracioso.

### Variáveis obrigatórias

```text
FATURAMENTO_HTTP_PORT
FATURAMENTO_DATABASE_URL
FATURAMENTO_ESTOQUE_BASE_URL
FATURAMENTO_RABBITMQ_URL
FATURAMENTO_OUTBOX_INTERVAL
```

### Dependências e validação

PowerShell, Bash, Zsh, Git Bash ou Prompt de Comando:

```console
cd services/faturamento
go mod download
go test ./...
go vet ./...
```

Com PostgreSQL ativo e migrations aplicadas:

```console
go test -tags=integration ./internal/infrastructure/repository -count=1 -v
```

### Migrations e execução

Execute dentro de `services/faturamento`:

```console
go run ./cmd/migrate up
go run ./cmd/migrate version
go run ./cmd/api
```

Os comandos adicionais seguem o mesmo padrão do Estoque:

```console
go run ./cmd/migrate down
go run ./cmd/migrate force 2
```

API local:

| Recurso | Endereço |
|---|---|
| Health check | <http://localhost:8082/health> |
| Swagger UI | <http://localhost:8082/swagger/index.html> |

Endpoints implementados:

| Método | Rota | Resultado |
|---|---|---|
| `POST` | `/notas-fiscais` | Cria uma nota aberta com seus itens |
| `PUT` | `/notas-fiscais/{id}` | Substitui cliente, endereço e itens enquanto a nota estiver `ABERTA` |
| `GET` | `/notas-fiscais` | Lista as notas por número decrescente |
| `GET` | `/notas-fiscais/{id}` | Consulta uma nota e seus itens |
| `POST` | `/notas-fiscais/{id}/fechamento` | Muda para `PROCESSANDO` e grava a Outbox |

A listagem de notas fiscais aceita `pagina`, `tamanhoPagina` e os filtros
opcionais `numero`, `status` e `nomeCliente`:

```text
GET /notas-fiscais?pagina=1&tamanhoPagina=20&status=ABERTA&nomeCliente=maria
```

As duas APIs retornam listagens com os itens e os metadados da página:

```json
{
  "success": true,
  "data": {
    "itens": [],
    "total": 0,
    "pagina": 1,
    "tamanhoPagina": 20,
    "totalPaginas": 0
  }
}
```

Body mínimo para criar uma nota:

```json
{
  "nomeCliente": "Maria da Silva",
  "enderecoCliente": "Rua das Flores, 100 - Curitiba/PR",
  "itens": [
    {
      "codigoProduto": "SKU-001",
      "quantidade": 2
    }
  ]
}
```

O Faturamento consulta o Estoque pelo código. O cliente da API não envia ID,
descrição nem valor: esses dados são copiados do produto ativo e preservados
como snapshot no item da nota.

Na atualização, o body segue o mesmo formato mínimo da criação. Os produtos são
consultados novamente no Estoque, precisam continuar ativos e seus snapshots são
renovados. A operação substitui todos os itens e retorna conflito (`409`) se a
nota já estiver `PROCESSANDO` ou `FECHADA`.

Ao iniciar a API, o worker consulta periodicamente os registros com
`published_at IS NULL`, publica mensagens persistentes na exchange durável
`korp.events` usando a routing key `estoque.baixa.solicitada` e somente marca o
evento como publicado depois da confirmação do RabbitMQ. Em caso de falha, ele
permanece pendente para uma nova tentativa.

O consumo da solicitação pelo Estoque e a publicação dos eventos de sucesso ou
rejeição são a próxima etapa da integração assíncrona.

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
│   ├── faturamento/
│   │   ├── cmd/api/
│   │   ├── cmd/migrate/
│   │   ├── docs/
│   │   ├── internal/
│   │   └── migrations/
│   └── estoque/
│       ├── cmd/api/
│       ├── cmd/migrate/
│       ├── docs/
│       ├── internal/dependency/
│       ├── internal/application/produto/
│       │   ├── criar_produto.go
│       │   ├── listar_produtos.go
│       │   ├── buscar_produto_por_id.go
│       │   ├── buscar_produto_por_codigo.go
│       │   ├── atualizar_produto.go
│       │   ├── ativar_produto.go
│       │   └── inativar_produto.go
│       ├── internal/application/estoque/
│       │   ├── baixar_estoque.go
│       │   └── listar_movimentacoes.go
│       ├── internal/infrastructure/database/models/
│       ├── internal/infrastructure/repository/
│       ├── internal/presentation/http/
│       │   ├── domainerror/
│       │   ├── dto/
│       │   └── response/
│       ├── internal/shared/text/
│       └── migrations/
├── .env.example
├── docker-compose.yml
└── README.md
```

## Documentação arquitetural

- [Arquitetura](docs/README_ARQUITETURA.md)
- [Decisões arquiteturais](docs/README_DECISOES_ARQUITETURAIS.md)
- [Modelo de domínio](docs/README_MODELO_DOMINIO.md)
