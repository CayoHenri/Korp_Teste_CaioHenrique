# Documentação técnica

Este diretório concentra o detalhamento da solução. Cada assunto possui uma
fonte principal para evitar comandos ou decisões duplicadas e contraditórias.

## Ordem sugerida de leitura

1. [Arquitetura](README_ARQUITETURA.md): componentes, camadas e fluxo completo;
2. [Decisões arquiteturais](README_DECISOES_ARQUITETURAIS.md): escolhas e consequências;
3. [Modelo de domínio](README_MODELO_DOMINIO.md): entidades, estados e invariantes;
4. [Estoque](SERVICO_ESTOQUE.md): API, configuração, migrations e regras do serviço;
5. [Faturamento](SERVICO_FATURAMENTO.md): API, configuração, Outbox e fechamento;
6. [Resiliência](RESILIENCIA.md): retry, DLQ, confirmação e idempotência;
7. [Docker](EXECUCAO_DOCKER.md): execução completa ou desenvolvimento local;
8. [Testes](TESTES.md): suítes, pré-requisitos e cenários cobertos.
9. [Frontend](../frontend/web/README.md): estrutura Angular, RxJS e Angular Material.

## Mapa por necessidade

| Preciso... | Documento |
|---|---|
| Entender o sistema | [Arquitetura](README_ARQUITETURA.md) |
| Saber por que uma tecnologia foi escolhida | [Decisões](README_DECISOES_ARQUITETURAIS.md) |
| Entender as regras de Produto e Nota Fiscal | [Domínio](README_MODELO_DOMINIO.md) |
| Consultar endpoints do Estoque | [Serviço de Estoque](SERVICO_ESTOQUE.md) |
| Consultar endpoints do Faturamento | [Serviço de Faturamento](SERVICO_FATURAMENTO.md) |
| Entender RabbitMQ, Outbox ou duplicidade | [Resiliência](RESILIENCIA.md) |
| Subir ou diagnosticar os containers | [Docker](EXECUCAO_DOCKER.md) |
| Executar ou entender os testes | [Testes](TESTES.md) |
| Entender ou evoluir o frontend | [Frontend](../frontend/web/README.md) |

## Contratos gerados

Cada API expõe seu contrato OpenAPI pelo Swagger UI e mantém os artefatos
gerados no diretório `docs` do próprio módulo Go:

- Estoque: `services/estoque/docs`;
- Faturamento: `services/faturamento/docs`.

Esses artefatos são contratos de API gerados; não substituem os documentos de
arquitetura deste diretório.
