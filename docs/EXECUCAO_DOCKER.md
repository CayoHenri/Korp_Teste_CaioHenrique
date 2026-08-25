# Execução local e Docker

## Decisão de execução

O Docker é usado somente para a infraestrutura e para aplicar migrations:

| Serviço | Tipo | Responsabilidade |
|---|---|---|
| `postgres` | contínuo | Banco dos schemas `estoque` e `faturamento` |
| `rabbitmq` | contínuo | Mensageria e interface administrativa |
| `estoque-migrations` | execução única | Aplicar o schema do Estoque |
| `faturamento-migrations` | execução única | Aplicar o schema do Faturamento |

As APIs Go e o frontend Angular são executados diretamente no computador. Essa
escolha reduz CPU, memória e tempo de build no Docker Desktop e mantém o ciclo
de desenvolvimento rápido.

## Pré-requisitos

- Docker Desktop com Docker Compose;
- Go 1.25;
- Node.js 24 e npm;
- portas `5432`, `5672`, `15672`, `8081`, `8082` e `4200` livres;
- `.env` criado na raiz a partir de `.env.example`.

PowerShell:

```powershell
Copy-Item .env.example .env
```

Bash, Zsh ou Git Bash:

```bash
cp .env.example .env
```

O código não possui fallback para variáveis obrigatórias. Os serviços procuram
o `.env` da raiz por busca ascendente, mesmo quando iniciados em suas pastas.

## 1. Subir infraestrutura e migrations

Execute na raiz. Os comandos funcionam em PowerShell e Bash:

```console
docker compose config
docker compose up -d postgres rabbitmq
docker compose build estoque-migrations
docker compose run --rm estoque-migrations up
docker compose build faturamento-migrations
docker compose run --rm faturamento-migrations up
docker compose ps -a
```

PostgreSQL e RabbitMQ permanecem ativos. As migrations são executadas com
`--rm`: o sucesso é indicado pelo código de saída `0` e o container temporário é
removido logo depois. Suas imagens compilam somente `cmd/migrate`; nenhuma API é
compilada ou executada pelo Compose.
Os builds aparecem separados de propósito para evitar duas compilações Go ao
mesmo tempo em máquinas com pouca memória disponível para o Docker Desktop.

Recompilar e reaplicar migrations:

```console
docker compose build estoque-migrations
docker compose run --rm estoque-migrations up
docker compose build faturamento-migrations
docker compose run --rm faturamento-migrations up
```

## 2. Executar a API de Estoque

Abra outro terminal.

PowerShell:

```powershell
Set-Location services/estoque
go mod download
go run ./cmd/api
```

Bash, Zsh ou Git Bash:

```bash
cd services/estoque
go mod download
go run ./cmd/api
```

Valide em <http://localhost:8081/health> e
<http://localhost:8081/swagger/index.html>.

## 3. Executar a API de Faturamento

Abra outro terminal, mantendo Estoque ativo.

PowerShell:

```powershell
Set-Location services/faturamento
go mod download
go run ./cmd/api
```

Bash, Zsh ou Git Bash:

```bash
cd services/faturamento
go mod download
go run ./cmd/api
```

Valide em <http://localhost:8082/health> e
<http://localhost:8082/swagger/index.html>.

## 4. Executar o frontend Angular

Abra um quarto terminal.

PowerShell:

```powershell
Set-Location frontend/web
npm install
npx ng serve
```

Bash, Zsh ou Git Bash:

```bash
cd frontend/web
npm install
npx ng serve
```

Acesse <http://localhost:4200>. O frontend chama as APIs em `8081` e `8082`;
as origens CORS correspondentes estão configuradas no `.env.example`.
`npx ng` usa o Angular CLI instalado no workspace, garantindo a versão declarada
no projeto sem exigir instalação global.

## Migrations manuais sem Docker

Normalmente o Compose já aplica migrations. Para controlar versão, rollback ou
force, execute na pasta do respectivo serviço:

```console
go run ./cmd/migrate up
go run ./cmd/migrate version
go run ./cmd/migrate down
go run ./cmd/migrate force NUMERO_DA_VERSAO
```

`down` reverte uma migration. `force` apenas altera a versão registrada e não
corrige o schema; use somente após inspecionar o banco.

## Dados iniciais de demonstração

As migrations aplicam também um conjunto determinístico para uso local:

- Estoque `000006`: 14 produtos de informática, sendo 12 ativos e 2 inativos,
  com preços, saldos e movimentações iniciais coerentes;
- Faturamento `000008`: 5 notas fiscais abertas, 12 itens e clientes/endereços
  fictícios, referenciando os mesmos produtos do Estoque;
- a sequência de número das notas continua depois de `1005`.

Esses dados fazem parte das migrations para que um banco criado do zero esteja
pronto para navegar e testar. Os arquivos `down` removem somente os registros
determinísticos dos respectivos seeds.

## Endereços

| Recurso | Endereço local |
|---|---|
| Frontend | `http://localhost:4200` |
| Estoque | `http://localhost:8081` |
| Faturamento | `http://localhost:8082` |
| PostgreSQL | `localhost:5432` |
| RabbitMQ AMQP | `localhost:5672` |
| RabbitMQ Management | `http://localhost:15672` |

## Operação e diagnóstico

```console
docker compose ps -a
docker compose logs -f postgres rabbitmq
docker compose down
```

`docker compose down` preserva os volumes. A variante `--volumes` apaga todos os
dados locais de PostgreSQL e RabbitMQ e não deve ser usada apenas para reiniciar.

Para recriar intencionalmente todo o ambiente de desenvolvimento:

```console
docker compose down --volumes
```

Depois repita a seção “Subir infraestrutura e migrations”. Essa ação é
irreversível para os dados armazenados apenas nesses volumes.

Se uma API não iniciar, confirme nesta ordem:

1. PostgreSQL e RabbitMQ estão `healthy` em `docker compose ps`;
2. os comandos de migration terminaram com código `0`;
3. o `.env` está na raiz e contém todas as variáveis;
4. a porta da API está livre;
5. o terminal está na pasta correta do serviço.
