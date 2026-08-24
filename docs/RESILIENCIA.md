# Resiliência do processamento assíncrono

## Objetivo

Esta solução busca evitar perda de mensagens e processamento duplicado sem
adicionar mecanismos mais complexos do que o necessário para o projeto.

O fluxo assíncrono é usado quando o Faturamento solicita uma baixa ao Estoque.
O Faturamento não espera a baixa terminar durante a requisição HTTP: ele grava
um evento e o processamento continua pelos workers e pelo RabbitMQ.

## Visão geral do fluxo

1. O Faturamento muda a nota para `PROCESSANDO`.
2. Na mesma transação, grava um evento na tabela Outbox.
3. O worker da Outbox publica a solicitação no RabbitMQ.
4. O Estoque consome a solicitação e tenta realizar a baixa.
5. O Estoque publica o resultado de sucesso ou rejeição.
6. O Faturamento consome o resultado e atualiza a nota fiscal.

## Decisões adotadas

### Outbox transacional

A mudança de estado da nota e a criação do evento acontecem na mesma transação
do PostgreSQL. Portanto, as duas operações são confirmadas juntas ou desfeitas
juntas.

Isso evita o problema em que a nota fica como `PROCESSANDO`, mas a aplicação
falha antes de criar a mensagem que solicita a baixa.

O worker busca eventos com `published_at` vazio. Depois que o RabbitMQ confirma
a publicação, o worker preenche `published_at`.

### Confirmação manual

Os consumidores não confirmam uma mensagem assim que a recebem. O `Ack` é
enviado somente depois que o processamento ou o encaminhamento da mensagem foi
concluído.

Se a própria tentativa de publicar no retry ou na DLQ falhar, é usado `Nack`
com reentrega. Assim, a mensagem original não é descartada silenciosamente.

### Idempotência

RabbitMQ e Outbox trabalham com garantia de entrega de pelo menos uma vez. Isso
significa que uma mensagem pode chegar novamente, por exemplo, quando o serviço
processa a operação, mas perde a conexão antes de confirmar o `Ack`.

Para impedir efeitos duplicados, os identificadores das mensagens processadas
são persistidos. Uma mensagem repetida é reconhecida e não baixa o estoque nem
altera a nota novamente.

### DLQ para mensagens inválidas

Mensagens com JSON inválido, campos obrigatórios ausentes ou erros terminais
são enviadas para uma Dead Letter Queue (DLQ). Isso impede que uma mensagem que
nunca poderá ser processada fique circulando indefinidamente.

As filas são:

- `estoque.baixa.solicitada.dlq`;
- `faturamento.baixa.resultado.dlq`.

### Reconexão automática

Os clientes AMQP tentam restabelecer a conexão com o RabbitMQ quando ocorre uma
interrupção. A quantidade de tentativas e o intervalo são configurados por:

- `RABBITMQ_RECOVERY_MAX_RETRIES`;
- `RABBITMQ_RECOVERY_INTERVAL`.

### Retry simples e limitado

Falhas técnicas temporárias, como indisponibilidade momentânea do banco, são
enviadas para uma fila de retry. A mensagem espera um intervalo fixo antes de
voltar à fila principal.

As configurações são:

- `RABBITMQ_MESSAGE_MAX_RETRIES`: limite de tentativas;
- `RABBITMQ_MESSAGE_RETRY_DELAY`: tempo fixo entre tentativas;
- `RABBITMQ_MESSAGE_TIMEOUT`: tempo máximo de cada processamento ou publicação.

Quando o limite é atingido, a mensagem segue para a DLQ. Isso evita um loop
infinito consumindo CPU e gerando logs repetidos.

O Estoque possui uma fila de retry para solicitações de baixa. O Faturamento
possui uma fila de retry para os resultados, independentemente de o resultado
ser sucesso ou rejeição.

## Simplificações conscientes

O projeto considera uma única instância do worker da Outbox. Por isso, não usa
claim concorrente, `FOR UPDATE SKIP LOCKED`, `locked_at` ou `locked_by`.

Também foi escolhido atraso fixo em vez de backoff exponencial. Essa opção é
mais fácil de entender, configurar e testar, e atende ao escopo atual.

Não existe teste automatizado que derruba o container do RabbitMQ, pois esse
tipo de teste altera o ambiente compartilhado. A reconexão pode ser conferida
manualmente em um ambiente isolado.

Se no futuro houver várias instâncias da Outbox ou um volume muito maior de
mensagens, poderão ser avaliados claim concorrente, backoff exponencial e testes
de falha controlada.

## Garantia oferecida

A solução oferece processamento **pelo menos uma vez**, combinado com
idempotência. Ela não promete entrega exatamente uma vez, porque pode existir
uma falha entre a confirmação do RabbitMQ e a atualização de `published_at`.

Nesse caso, a mensagem pode ser publicada novamente, mas os controles de
idempotência impedem que o efeito de negócio seja repetido.
