# Decisões Arquiteturais

## 1. Objetivo

Este documento registra as principais decisões arquiteturais tomadas para o projeto de emissão de notas fiscais.

O objetivo é deixar explícito:

- o contexto de cada decisão;
- a alternativa escolhida;
- as alternativas rejeitadas;
- as consequências da decisão;
- os possíveis caminhos de evolução.

O formato segue uma abordagem simplificada de **Architecture Decision Records (ADR)**.

---

# ADR-001 — Utilizar dois microsserviços

## Status

Aceita.

## Contexto

O desafio exige no mínimo:

- Serviço de Estoque;
- Serviço de Faturamento.

Os domínios possuem responsabilidades diferentes.

## Decisão

Criar dois microsserviços independentes:

```text
estoque-service
faturamento-service
```

## Responsabilidades

### Estoque

- produtos;
- saldos;
- movimentações;
- baixa de estoque;
- concorrência;
- idempotência da baixa.

### Faturamento

- notas fiscais;
- numeração;
- itens;
- estado da nota;
- fechamento;
- coordenação da emissão.

## Consequências positivas

- separação de responsabilidades;
- baixo acoplamento;
- maior clareza do domínio;
- facilita demonstrar arquitetura de microsserviços.

## Consequências negativas

- aumenta complexidade de integração;
- exige lidar com falhas parciais;
- exige consistência eventual.

---

# ADR-002 — Utilizar Go no backend

## Status

Aceita.

## Contexto

O desafio permite backend em Go ou C#.

## Decisão

Utilizar Go.

## Motivos

- baixo overhead;
- bom suporte para concorrência;
- fácil containerização;
- ecossistema maduro para HTTP;
- excelente integração com PostgreSQL;
- bom suporte para RabbitMQ;
- tratamento explícito de erros.

## Framework HTTP

Preferência:

```text
Gin
```

## Consequências

A aplicação continuará estruturada para não depender do Gin no domínio.

---

# ADR-003 — Utilizar Angular no frontend

## Status

Aceita.

## Contexto

Angular é requisito explícito do desafio.

## Decisão

Utilizar Angular com:

- Reactive Forms;
- HttpClient;
- RxJS;
- Angular Material.

## Justificativa

Essa combinação permite demonstrar claramente os pontos solicitados na apresentação técnica:

- ciclos de vida;
- uso de RxJS;
- bibliotecas adicionais;
- biblioteca visual.

---

# ADR-004 — Utilizar PostgreSQL

## Status

Aceita.

## Contexto

O desafio exige conexão real com banco.

## Decisão

Utilizar PostgreSQL.

## Motivos

- banco relacional consolidado;
- bom suporte a transações;
- constraints;
- JSONB;
- UUID;
- locking;
- concorrência;
- excelente suporte em Go;
- simples de executar via Docker.

---

# ADR-005 — Compartilhar a mesma instância física do PostgreSQL

## Status

Aceita para o desafio técnico.

## Contexto

Uma implantação completa de microsserviços normalmente tende a utilizar isolamento físico maior de dados.

Entretanto, para um desafio técnico pequeno, manter duas instâncias PostgreSQL aumenta a infraestrutura sem necessariamente agregar valor proporcional.

## Decisão

Usar uma única instância física:

```text
postgres
```

com um banco:

```text
korp_db
```

e dois schemas:

```text
estoque
faturamento
```

## Regra

Cada microsserviço é proprietário de seu schema.

O Faturamento não manipula tabelas de Estoque.

O Estoque não manipula tabelas de Faturamento.

## Consequência

Existe isolamento lógico, mas não isolamento físico.

## Evolução futura

Migrar facilmente para:

```text
estoque_db
faturamento_db
```

sem alterar o modelo de comunicação entre os serviços.

---

# ADR-006 — Utilizar RabbitMQ

## Status

Aceita.

## Contexto

O fechamento de uma nota exige baixa de estoque.

Poderíamos usar HTTP síncrono ou mensageria.

## Decisão

Utilizar RabbitMQ para comunicação entre Faturamento e Estoque no fluxo de fechamento.

## Motivos

- desacoplamento temporal;
- possibilidade de recuperação após indisponibilidade;
- filas persistentes;
- acknowledgements;
- retries;
- suporte a DLQ;
- simples execução via Docker.

## Alternativa rejeitada

HTTP síncrono como única forma de integração.

### Problema

Se Estoque estiver indisponível:

```text
Faturamento -> Estoque -> falha
```

a operação inteira precisa ser tratada imediatamente.

Com RabbitMQ, a mensagem pode aguardar o retorno do consumidor.

---

# ADR-007 — Utilizar comunicação HTTP e mensageria

## Status

Aceita.

## Decisão

Nem toda comunicação utilizará RabbitMQ.

### HTTP

Utilizado para operações que precisam de resposta direta ao frontend:

```text
cadastrar produto
listar produto
criar nota
consultar nota
```

### RabbitMQ

Utilizado para integração assíncrona:

```text
solicitar baixa
confirmar baixa
rejeitar baixa
```

## Motivo

Evita aplicar mensageria onde ela não traz benefício real.

---

# ADR-008 — Não permitir acesso cruzado entre schemas

## Status

Aceita.

## Contexto

Como os serviços usam a mesma instância PostgreSQL, tecnicamente seria possível um serviço consultar as tabelas do outro.

## Decisão

Proibir esse acoplamento.

## Exemplo não permitido

```sql
SELECT *
FROM estoque.produtos;
```

executado pelo Faturamento.

## Forma correta

```text
Faturamento
    |
    +-- API ou evento
           |
           v
        Estoque
```

## Motivo

Preserva a autonomia dos microsserviços.

---

# ADR-009 — Utilizar estado interno PROCESSANDO

## Status

Aceita.

## Contexto

O desafio define:

- Aberta;
- Fechada.

Entretanto, a comunicação assíncrona cria um intervalo entre solicitação e conclusão.

## Decisão

Utilizar internamente:

```text
ABERTA
PROCESSANDO
FECHADA
```

## Fluxo

```text
ABERTA
   |
   v
PROCESSANDO
   |
   +-- sucesso --> FECHADA
   |
   +-- erro ----> ABERTA
```

## Motivo

Permite:

- impedir dupla solicitação;
- exibir loading corretamente;
- representar operação assíncrona;
- identificar notas presas.

---

# ADR-010 — Utilizar Transactional Outbox

## Status

Aceita.

## Contexto

Salvar no PostgreSQL e publicar no RabbitMQ são operações independentes.

Existe risco de:

```text
commit banco OK
publicação RabbitMQ falha
```

## Decisão

Persistir o evento na mesma transação da nota.

Exemplo:

```text
BEGIN

UPDATE notas_fiscais ...
INSERT INTO outbox_events ...

COMMIT
```

Um worker publica posteriormente.

## Benefício

Reduz o risco de perda de mensagens.

---

# ADR-011 — Implementar idempotência

## Status

Aceita.

## Contexto

Mensageria normalmente opera com entrega "at least once".

Isso significa que uma mensagem pode ser consumida mais de uma vez.

## Problema

Sem idempotência:

```text
saldo 10

mensagem 1:
10 -> 8

mesma mensagem novamente:
8 -> 6
```

## Decisão

Registrar IDs processados.

```text
estoque.mensagens_processadas
```

## Resultado

A mesma mensagem não produz um segundo efeito.

---

# ADR-012 — Controle de concorrência no PostgreSQL

## Status

Aceita.

## Contexto

O desafio apresenta como cenário opcional:

```text
saldo = 1

nota A usa 1
nota B usa 1
```

## Decisão

Utilizar atualização atômica:

```sql
UPDATE estoque.produtos
SET saldo = saldo - $1
WHERE id = $2
  AND saldo >= $1;
```

## Interpretação

```text
RowsAffected = 1
    sucesso

RowsAffected = 0
    saldo insuficiente
```

## Benefício

Evita saldo negativo.

---

# ADR-013 — Utilizar monorepo

## Status

Aceita.

## Contexto

Embora os serviços sejam separados logicamente, o projeto será avaliado como uma única entrega.

## Decisão

Utilizar monorepo.

```text
Korp_Teste_CaioHenrique/
|
+-- frontend/
+-- services/
|   +-- estoque/
|   +-- faturamento/
+-- infrastructure/
+-- docs/
```

## Motivos

- facilidade de clonagem;
- facilidade de avaliação;
- Docker Compose centralizado;
- documentação única;
- entrega simples.

---

# ADR-014 — Cada serviço possui seu próprio módulo Go

## Status

Aceita.

## Decisão

Cada serviço possuirá seu próprio:

```text
go.mod
go.sum
```

Exemplo:

```text
services/estoque/go.mod
services/faturamento/go.mod
```

## Benefícios

- dependências independentes;
- maior isolamento;
- possibilidade de deploy independente;
- evita acoplamento acidental.

---

# ADR-015 — Estrutura interna inspirada em Clean Architecture

## Status

Aceita.

## Decisão

Utilizar:

```text
domain
application
infrastructure
presentation
```

## Responsabilidades

### domain

Regras de negócio.

### application

Casos de uso e orquestração.

### infrastructure

PostgreSQL, RabbitMQ, clientes externos.

### presentation

HTTP, handlers, requests e responses.

## Observação

A estrutura será pragmática.

Não será criada abstração apenas por abstração.

---

# ADR-016 — Angular Material para componentes visuais

## Status

Aceita.

## Decisão

Utilizar Angular Material.

## Componentes esperados

```text
MatTable
MatFormField
MatInput
MatButton
MatSelect
MatIcon
MatDialog
MatSnackBar
MatProgressSpinner
```

## Motivo

Permite construir rapidamente uma UI consistente e profissional.

---

# ADR-017 — RxJS para operações assíncronas

## Status

Aceita.

## Aplicações

- requisições HTTP;
- loading;
- tratamento de erro;
- composição de streams;
- `finalize`;
- `switchMap` quando apropriado;
- cancelamento automático de subscriptions.

## Exemplo

```typescript
this.service.fechar(id)
  .pipe(
    finalize(() => this.loading = false)
  )
  .subscribe(...);
```

---

# ADR-018 — Docker Compose para ambiente local

## Status

Aceita.

## Decisão

A infraestrutura será executável com Docker Compose.

Serviços previstos:

```text
postgres
rabbitmq
estoque-service
faturamento-service
frontend
```

## Benefício

O avaliador consegue subir o projeto com poucos comandos.

---

# ADR-019 — Não utilizar Kafka

## Status

Rejeitada.

## Motivo

Kafka seria tecnicamente possível, mas adicionaria complexidade desnecessária para o tamanho do projeto.

RabbitMQ atende melhor ao caso de:

- comandos;
- filas;
- processamento assíncrono;
- retries;
- mensagens de integração.

---

# ADR-020 — IA não será prioridade

## Status

Adiada.

## Contexto

IA é opcional.

## Decisão

Priorizar:

1. microsserviços;
2. falhas;
3. concorrência;
4. idempotência;
5. qualidade arquitetural.

## Motivo

Esses itens possuem relação direta com o problema de negócio e demonstram melhor maturidade técnica.

---

# ADR-021 — Utilizar GORM para persistência

## Status

Aceita.

## Decisão

Utilizar GORM nos repositórios, consultas e transações dos microsserviços Go.

O domínio e os casos de uso não dependerão diretamente do GORM. Os models e a
implementação concreta dos repositórios permanecerão na infraestrutura.

## Restrição

Não utilizar `AutoMigrate`. A evolução do schema deve ser explícita e versionada.

---

# ADR-022 — Utilizar golang-migrate

## Status

Aceita.

## Decisão

Cada microsserviço possuirá migrations SQL `up` e `down` próprias, controladas
por um comando separado baseado em `golang-migrate`.

## Motivos

- histórico de alterações auditável;
- aplicação e reversão controladas;
- separação entre inicialização da API e alteração de schema;
- compatibilidade com execução local, CI e deploy.

---

# ADR-023 — Documentar APIs com Swagger/OpenAPI

## Status

Aceita.

## Decisão

Gerar a especificação Swagger a partir de anotações nos endpoints Gin e expor
uma interface Swagger UI em cada microsserviço.

Os artefatos gerados serão versionados para que o contrato possa ser consultado
sem exigir o gerador durante a execução da aplicação.

---

# ADR-024 — Exigir configuração explícita

## Status

Aceita.

## Decisão

Variáveis necessárias não possuirão fallback no código nem no Docker Compose.
A aplicação deve falhar imediatamente quando uma configuração obrigatória
estiver ausente ou vazia.

O carregamento e a validação de configuração pertencem à infraestrutura.

## Motivo

Evitar conexões silenciosas com banco, porta ou credenciais incorretas e tornar
o contrato de execução explícito por meio do `.env.example`.

---

# Resumo das decisões

| Decisão | Escolha |
|---|---|
| Frontend | Angular |
| UI | Angular Material |
| Backend | Go |
| HTTP | Gin |
| Persistência Go | GORM |
| Evolução do schema | golang-migrate |
| Documentação HTTP | Swagger/OpenAPI |
| Configuração | Variáveis obrigatórias, sem fallback |
| Banco | PostgreSQL |
| Separação de dados | Schemas |
| Mensageria | RabbitMQ |
| Comunicação crítica | Assíncrona |
| Persistência de eventos | Transactional Outbox |
| Idempotência | Event ID |
| Concorrência | UPDATE atômico |
| Organização | Monorepo |
| Infra local | Docker Compose |
| Arquitetura interna | Clean Architecture pragmática |

---

# Princípio principal

A principal diretriz da arquitetura é:

> Mesmo compartilhando infraestrutura física, cada microsserviço continua sendo proprietário de seus dados e de suas regras de negócio.

Isso permite manter uma solução simples para o desafio sem abandonar os princípios fundamentais de uma arquitetura distribuída.
