# Serviço de Estoque

## Responsabilidade

O Estoque é o único proprietário de produtos, saldos e movimentações. Nenhum
outro serviço consulta diretamente o schema `estoque`.

Suas responsabilidades são:

- cadastrar, consultar e atualizar produtos;
- ativar e inativar produtos sem exclusão física;
- registrar toda alteração de saldo;
- processar baixas de notas de forma transacional;
- impedir saldo negativo sob concorrência;
- ignorar solicitações repetidas;
- publicar o resultado da baixa.

## Organização interna

```text
internal/
├── application/
│   ├── produto/             um use case por operação cadastral
│   └── estoque/             baixa e consulta de movimentações
├── dependency/              composição das dependências
├── domain/
│   ├── produto/
│   └── movimentacao/
├── infrastructure/
│   ├── config/
│   ├── database/models/     models GORM e ToDomain
│   ├── messaging/           RabbitMQ
│   └── repository/          persistência PostgreSQL
├── presentation/http/
│   ├── domainerror/         domínio para resposta HTTP
│   ├── dto/                 contratos de entrada e saída
│   └── response/            envelopes HTTP
└── shared/                  paginação, filtros e texto
```

O domínio não depende de Gin, GORM ou RabbitMQ. A infraestrutura implementa as
interfaces necessárias e `internal/dependency` injeta as implementações.

## Produto

Um produto possui código, descrição, saldo, valor e estado ativo/inativo.

Regras principais:

- código e descrição são obrigatórios e normalizados em uppercase;
- código é único e não pode ser alterado;
- saldo e valor não podem ser negativos;
- produto novo inicia ativo;
- produto não é excluído, preservando referências históricas;
- atualização cadastral permite descrição, saldo e valor;
- alteração de saldo gera uma movimentação de entrada ou saída.

As propriedades do domínio são privadas. Leitura ocorre por getters e alterações
por métodos que representam ações de negócio.

## Baixa de estoque

O consumidor recebe uma lista de produto e quantidade. O processamento ocorre
em uma transação única:

1. verifica se o `eventId` já foi processado;
2. valida todos os itens;
3. aplica atualizações atômicas de saldo;
4. registra as movimentações;
5. persiste o identificador processado;
6. confirma a transação;
7. publica sucesso ou rejeição.

Se qualquer item falhar, toda a transação é revertida. O saldo é atualizado com
condição equivalente a `saldo >= quantidade`, evitando saldo negativo quando
duas baixas concorrem pelo mesmo produto.

## API HTTP

| Método | Rota | Finalidade |
|---|---|---|
| `POST` | `/produtos` | Criar produto |
| `GET` | `/produtos` | Listar com paginação e filtros |
| `GET` | `/produtos/{id}` | Buscar por UUID |
| `GET` | `/produtos/codigo/{codigo}` | Buscar por código |
| `PUT` | `/produtos/{id}` | Atualizar campos permitidos |
| `PATCH` | `/produtos/{id}/ativar` | Ativar produto |
| `PATCH` | `/produtos/{id}/inativar` | Inativar produto |
| `GET` | `/produtos/{id}/movimentacoes` | Consultar histórico |
| `GET` | `/health` | Verificar API e banco |

### Paginação e filtros

```text
GET /produtos?pagina=1&tamanhoPagina=20&codigo=SKU&descricao=TECLADO&ativo=true
```

O tamanho máximo de página é 100. A resposta contém `itens`, `total`, `pagina`,
`tamanhoPagina` e `totalPaginas`. Os produtos são ordenados pela data de cadastro
decrescente; em caso de empate, o UUID em ordem decrescente estabiliza a paginação.

### Criar produto

```json
{
  "codigo": "SKU-001",
  "descricao": "Teclado mecânico",
  "saldo": 10,
  "valor": 159.90
}
```

PowerShell:

```powershell
$body = @{
    codigo = "SKU-001"
    descricao = "Teclado mecânico"
    saldo = 10
    valor = 159.90
} | ConvertTo-Json

Invoke-RestMethod `
    -Method Post `
    -Uri "http://localhost:8081/produtos" `
    -ContentType "application/json" `
    -Body $body
```

Bash, Zsh ou Git Bash:

```bash
curl -X POST http://localhost:8081/produtos \
  -H 'Content-Type: application/json' \
  -d '{"codigo":"SKU-001","descricao":"Teclado mecânico","saldo":10,"valor":159.90}'
```

## Respostas HTTP

Sucesso:

```json
{"success": true, "data": {}}
```

Erro conhecido:

```json
{
  "success": false,
  "error": {
    "code": "PRODUTO_NAO_ENCONTRADO",
    "message": "produto nao encontrado"
  }
}
```

Erros inesperados retornam `ERRO_INTERNO` sem detalhes de banco ou stack trace.

## Configuração obrigatória

```text
ESTOQUE_HTTP_PORT
ESTOQUE_CORS_ALLOWED_ORIGINS
ESTOQUE_DATABASE_URL
ESTOQUE_RABBITMQ_URL
RABBITMQ_RECOVERY_MAX_RETRIES
RABBITMQ_RECOVERY_INTERVAL
RABBITMQ_MESSAGE_TIMEOUT
RABBITMQ_MESSAGE_MAX_RETRIES
RABBITMQ_MESSAGE_RETRY_DELAY
```

Não existem fallbacks. O serviço encerra imediatamente se uma variável estiver
ausente ou inválida. Em execução local, o `.env` da raiz é encontrado pela busca
ascendente da infraestrutura de configuração.

## Execução local

Com PostgreSQL e RabbitMQ ativos:

```console
cd services/estoque
go mod download
go run ./cmd/migrate up
go run ./cmd/api
```

Swagger: <http://localhost:8081/swagger/index.html>.

## Migrations

Execute dentro de `services/estoque`, pois `file://migrations` é relativo ao
diretório atual:

```console
go run ./cmd/migrate up
go run ./cmd/migrate version
go run ./cmd/migrate down
go run ./cmd/migrate force 5
```

`down` reverte uma migration. `force` apenas corrige a versão registrada e deve
ser usado após conferir manualmente o banco; ele não executa SQL.

Uma migration nova exige arquivos `.up.sql` e `.down.sql` com o próximo número
sequencial. O projeto não usa `AutoMigrate`.

A migration `000006_seed_produtos_realistas` fornece dados de demonstração para
ambientes criados do zero: produtos ativos e inativos, valores, saldos e uma
movimentação de entrada correspondente para cada produto. IDs e códigos são
determinísticos para manter consistência com os itens iniciais do Faturamento.

## Swagger

Após mudar anotações dos handlers:

```console
cd services/estoque
go run github.com/swaggo/swag/cmd/swag@v1.16.6 init -g cmd/api/main.go -o docs
```
