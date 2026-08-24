package messaging

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"sync"
	"time"

	application "github.com/caiog/korp-notas-fiscais/services/estoque/internal/application/estoque"
	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	EventsExchange       = "korp.events"
	DeadLetterExchange   = "korp.events.dlx"
	BaixaSolicitadaQueue = "estoque.baixa.solicitada"
	BaixaSolicitadaDLQ   = "estoque.baixa.solicitada.dlq"
	BaixaSolicitadaRetry = "estoque.baixa.solicitada.retry"
	ResultadoBaixaQueue  = "faturamento.baixa.resultado"
)

var ErrPublicacaoNaoConfirmada = errors.New("publicacao nao confirmada pelo RabbitMQ")

type RabbitMQ struct {
	connection        *amqp.Connection
	consumerChannel   *amqp.Channel
	publisherChannel  *amqp.Channel
	confirmations     <-chan amqp.Confirmation
	publisherMutex    sync.Mutex
	messageTimeout    time.Duration
	messageMaxRetries int
	messageRetryDelay time.Duration
}

func NewRabbitMQ(
	url string,
	maxRetries int,
	retryInterval, messageTimeout time.Duration,
	messageMaxRetries int,
	messageRetryDelay time.Duration,
) (*RabbitMQ, error) {
	connection, err := amqp.DialConfig(url, amqp.Config{Recovery: &amqp.Recovery{
		ReconnectionConfig: &amqp.ReconnectionConfig{
			MaxRetryCount: maxRetries,
			RetryInterval: retryInterval,
		},
	}})
	if err != nil {
		return nil, fmt.Errorf("conectar ao RabbitMQ: %w", err)
	}
	consumerChannel, err := connection.Channel()
	if err != nil {
		_ = connection.Close()
		return nil, fmt.Errorf("abrir canal consumidor: %w", err)
	}
	publisherChannel, err := connection.Channel()
	if err != nil {
		_ = consumerChannel.Close()
		_ = connection.Close()
		return nil, fmt.Errorf("abrir canal publisher: %w", err)
	}
	client := &RabbitMQ{
		connection:        connection,
		consumerChannel:   consumerChannel,
		publisherChannel:  publisherChannel,
		messageTimeout:    messageTimeout,
		messageMaxRetries: messageMaxRetries,
		messageRetryDelay: messageRetryDelay,
	}
	if err := client.configureTopology(); err != nil {
		_ = client.Close()
		return nil, err
	}
	if err := publisherChannel.Confirm(false); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("habilitar confirmacao de publicacao: %w", err)
	}
	client.confirmations = publisherChannel.NotifyPublish(make(chan amqp.Confirmation, 1))
	return client, nil
}

func (client *RabbitMQ) configureTopology() error {
	for _, exchange := range []string{EventsExchange, DeadLetterExchange} {
		if err := client.consumerChannel.ExchangeDeclare(
			exchange, "topic", true, false, false, false, nil,
		); err != nil {
			return fmt.Errorf("declarar exchange %s: %w", exchange, err)
		}
	}
	if _, err := client.consumerChannel.QueueDeclare(
		BaixaSolicitadaQueue, true, false, false, false, nil,
	); err != nil {
		return fmt.Errorf("declarar fila de solicitacao: %w", err)
	}
	if err := client.consumerChannel.QueueBind(
		BaixaSolicitadaQueue,
		application.EventTypeBaixaSolicitada,
		EventsExchange,
		false,
		nil,
	); err != nil {
		return fmt.Errorf("vincular fila de solicitacao: %w", err)
	}
	if _, err := client.consumerChannel.QueueDeclare(
		BaixaSolicitadaDLQ, true, false, false, false, nil,
	); err != nil {
		return fmt.Errorf("declarar DLQ: %w", err)
	}
	if _, err := client.consumerChannel.QueueDeclare(
		BaixaSolicitadaRetry, true, false, false, false, amqp.Table{
			"x-dead-letter-exchange":    EventsExchange,
			"x-dead-letter-routing-key": application.EventTypeBaixaSolicitada,
		},
	); err != nil {
		return fmt.Errorf("declarar fila de retentativa: %w", err)
	}
	if err := client.consumerChannel.QueueBind(
		BaixaSolicitadaDLQ, BaixaSolicitadaQueue, DeadLetterExchange, false, nil,
	); err != nil {
		return fmt.Errorf("vincular DLQ: %w", err)
	}
	if _, err := client.consumerChannel.QueueDeclare(
		ResultadoBaixaQueue, true, false, false, false, nil,
	); err != nil {
		return fmt.Errorf("declarar fila de resultados: %w", err)
	}
	for _, routingKey := range []string{
		application.EventTypeBaixaRealizada,
		application.EventTypeBaixaRejeitada,
	} {
		if err := client.consumerChannel.QueueBind(
			ResultadoBaixaQueue, routingKey, EventsExchange, false, nil,
		); err != nil {
			return fmt.Errorf("vincular fila de resultados: %w", err)
		}
	}
	return client.consumerChannel.Qos(1, 0, false)
}

type BaixaSolicitadaMessage struct {
	EventID      uuid.UUID                    `json:"eventId"`
	NotaFiscalID uuid.UUID                    `json:"notaFiscalId"`
	Itens        []BaixaSolicitadaItemMessage `json:"itens"`
}

type BaixaSolicitadaItemMessage struct {
	ProdutoID  uuid.UUID `json:"produtoId"`
	Quantidade int       `json:"quantidade"`
}

func (client *RabbitMQ) ConsumirBaixas(
	ctx context.Context,
	useCase *application.ProcessarBaixaSolicitadaUseCase,
) error {
	deliveries, err := client.consumerChannel.ConsumeWithContext(
		ctx, BaixaSolicitadaQueue, "", false, false, false, false, nil,
	)
	if err != nil {
		return fmt.Errorf("iniciar consumidor de baixas: %w", err)
	}
	for delivery := range deliveries {
		messageContext, cancel := context.WithTimeout(ctx, client.messageTimeout)
		var message BaixaSolicitadaMessage
		if err := json.Unmarshal(delivery.Body, &message); err != nil {
			if err := client.publicarMensagemInvalida(messageContext, delivery); err != nil {
				_ = delivery.Nack(false, true)
				cancel()
				continue
			}
			_ = delivery.Ack(false)
			cancel()
			continue
		}
		input := application.BaixarEstoqueInput{
			EventID:      message.EventID,
			NotaFiscalID: message.NotaFiscalID,
			Itens:        make([]application.BaixarEstoqueItemInput, 0, len(message.Itens)),
		}
		for _, item := range message.Itens {
			input.Itens = append(input.Itens, application.BaixarEstoqueItemInput{
				ProdutoID: item.ProdutoID, Quantidade: item.Quantidade,
			})
		}
		if err := useCase.Execute(messageContext, input); err != nil {
			if err := client.reagendarOuEnviarParaDLQ(ctx, delivery, BaixaSolicitadaRetry, BaixaSolicitadaQueue); err != nil {
				_ = delivery.Nack(false, true)
			} else {
				_ = delivery.Ack(false)
			}
			cancel()
			continue
		}
		_ = delivery.Ack(false)
		cancel()
	}
	return ctx.Err()
}

func (client *RabbitMQ) reagendarOuEnviarParaDLQ(
	ctx context.Context,
	delivery amqp.Delivery,
	retryQueue, deadLetterKey string,
) error {
	retry := retryCount(delivery.Headers)
	message := copiarMensagem(delivery)
	if retry >= client.messageMaxRetries {
		return client.publicarComConfirmacao(ctx, DeadLetterExchange, deadLetterKey, message)
	}
	if message.Headers == nil {
		message.Headers = amqp.Table{}
	}
	message.Headers["x-retry-count"] = int32(retry + 1)
	message.Expiration = fmt.Sprintf("%d", client.messageRetryDelay.Milliseconds())
	return client.publicarComConfirmacao(ctx, "", retryQueue, message)
}

func copiarMensagem(delivery amqp.Delivery) amqp.Publishing {
	headers := amqp.Table{}
	maps.Copy(headers, delivery.Headers)
	return amqp.Publishing{
		Headers:       headers,
		ContentType:   delivery.ContentType,
		DeliveryMode:  amqp.Persistent,
		MessageId:     delivery.MessageId,
		CorrelationId: delivery.CorrelationId,
		Timestamp:     delivery.Timestamp,
		Type:          delivery.Type,
		Body:          delivery.Body,
	}
}

func retryCount(headers amqp.Table) int {
	if headers == nil {
		return 0
	}
	switch value := headers["x-retry-count"].(type) {
	case int:
		return value
	case int32:
		return int(value)
	case int64:
		return int(value)
	default:
		return 0
	}
}

func (client *RabbitMQ) publicarMensagemInvalida(ctx context.Context, delivery amqp.Delivery) error {
	message := amqp.Publishing{
		Headers:       delivery.Headers,
		ContentType:   delivery.ContentType,
		DeliveryMode:  amqp.Persistent,
		MessageId:     delivery.MessageId,
		CorrelationId: delivery.CorrelationId,
		Timestamp:     time.Now().UTC(),
		Type:          delivery.Type,
		Body:          delivery.Body,
	}
	return client.publicarComConfirmacao(ctx, DeadLetterExchange, BaixaSolicitadaQueue, message)
}

func (client *RabbitMQ) PublicarResultado(ctx context.Context, result application.ResultadoBaixa) error {
	payload, err := json.Marshal(struct {
		EventID       uuid.UUID `json:"eventId"`
		CorrelationID uuid.UUID `json:"correlationId"`
		NotaFiscalID  uuid.UUID `json:"notaFiscalId"`
		Motivo        string    `json:"motivo,omitempty"`
	}{result.EventID, result.CorrelationID, result.NotaFiscalID, result.Motivo})
	if err != nil {
		return err
	}
	return client.publicarComConfirmacao(
		ctx,
		EventsExchange,
		result.Type,
		amqp.Publishing{
			ContentType:   "application/json",
			DeliveryMode:  amqp.Persistent,
			MessageId:     result.EventID.String(),
			CorrelationId: result.CorrelationID.String(),
			Timestamp:     time.Now().UTC(),
			Type:          result.Type,
			Body:          payload,
		},
	)
}

func (client *RabbitMQ) publicarComConfirmacao(
	ctx context.Context,
	exchange, routingKey string,
	message amqp.Publishing,
) error {
	requestContext, cancel := context.WithTimeout(ctx, client.messageTimeout)
	defer cancel()
	client.publisherMutex.Lock()
	defer client.publisherMutex.Unlock()
	if err := client.publisherChannel.PublishWithContext(
		requestContext, exchange, routingKey, false, false, message,
	); err != nil {
		return err
	}
	select {
	case confirmation, ok := <-client.confirmations:
		if !ok || !confirmation.Ack {
			return ErrPublicacaoNaoConfirmada
		}
		return nil
	case <-requestContext.Done():
		return requestContext.Err()
	}
}

func (client *RabbitMQ) Close() error {
	return errors.Join(
		client.consumerChannel.Close(),
		client.publisherChannel.Close(),
		client.connection.Close(),
	)
}
