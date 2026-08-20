# Sistema de Emissao de Notas Fiscais

Projeto de estudo baseado em Angular, microsservicos Go, PostgreSQL e RabbitMQ.

## Infraestrutura local

### Pre-requisitos

- Docker Desktop com Docker Compose;
- portas `5432`, `5672` e `15672` livres.

### Iniciar

Os valores padrao do ambiente de desenvolvimento ja estao definidos no
`docker-compose.yml`. Portanto, o arquivo `.env` e opcional.

```powershell
docker compose up -d
```

Para personalizar portas ou credenciais:

```powershell
Copy-Item .env.example .env
docker compose up -d
```

### Verificar

```powershell
docker compose ps
docker compose logs -f postgres rabbitmq
```

Servicos disponibilizados:

| Servico | Endereco | Uso |
|---|---|---|
| PostgreSQL | `localhost:5432` | Banco `korp_db` |
| RabbitMQ | `localhost:5672` | Conexao AMQP das aplicacoes |
| RabbitMQ Management | <http://localhost:15672> | Interface administrativa |

Credenciais padrao locais:

```text
usuario: korp
senha: korp_dev_password
```

> As credenciais padrao sao apenas para desenvolvimento local. Em outros
> ambientes, defina valores seguros por variaveis de ambiente.

### Parar

Para parar os containers e preservar os dados:

```powershell
docker compose down
```

Para remover tambem os volumes e reinicializar todos os dados:

```powershell
docker compose down --volumes
```

O script `infrastructure/postgres/init/001-create-schemas.sql` cria os schemas
`estoque` e `faturamento`. Scripts dentro de `docker-entrypoint-initdb.d` rodam
somente quando o volume do PostgreSQL ainda esta vazio.

## Documentacao

- [Arquitetura](docs/README_ARQUITETURA.md)
- [Decisoes arquiteturais](docs/README_DECISOES_ARQUITETURAIS.md)
- [Modelo de dominio](docs/README_MODELO_DOMINIO.md)
