package messaging

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	notafiscalApplication "github.com/caiog/korp-notas-fiscais/services/faturamento/internal/application/nota_fiscal"
	application "github.com/caiog/korp-notas-fiscais/services/faturamento/internal/application/outbox"
	notafiscal "github.com/caiog/korp-notas-fiscais/services/faturamento/internal/domain/nota_fiscal"
	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	EventsExchange       = "korp.events"
	DeadLetterExchange   = "korp.events.dlx"
	BaixaSolicitadaQueue = "estoque.baixa.solicitada"
	ResultadoBaixaQueue  = "faturamento.baixa.resultado"
	ResultadoBaixaDLQ    = "faturamento.baixa.resultado.dlq"
	ResultadoBaixaRetry  = "faturamento.baixa.resultado.retry"
	ResultadoRetryRoute  = "faturamento.baixa.resultado.retry"
)

var ErrPublicacaoNaoConfirmada = errors.New("publicacao nao confirmada pelo RabbitMQ")

type RabbitMQPublisher struct {
	connection        *amqp.Connection
	channel           *amqp.Channel
	consumerChannel   *amqp.Channel
	confirmations     <-chan amqp.Confirmation
	mutex             sync.Mutex
	messageTimeout    time.Duration
	messageMaxRetries int
	messageRetryDelay time.Duration
}

type resultadoBaixaMessage struct {
	EventID       uuid.UUID `json:"eventId"`
	CorrelationID uuid.UUID `json:"correlationId"`
	NotaFiscalID  uuid.UUID `json:"notaFiscalId"`
	Motivo        string    `json:"motivo,omitempty"`
}

func NewRabbitMQPublisher(
	url string,
	maxRetries int,
	retryInterval, messageTimeout time.Duration,
	messageMaxRetries int,
	messageRetryDelay time.Duration,
) (*RabbitMQPublisher, error) {
	connection, err := amqp.DialConfig(url, amqp.Config{Recovery: &amqp.Recovery{
		ReconnectionConfig: &amqp.ReconnectionConfig{
			MaxRetryCount: maxRetries,
			RetryInterval: retryInterval,
		},
	}})
	if err != nil {
		return nil, fmt.Errorf("conectar ao RabbitMQ: %w", err)
	}
	channel, err := connection.Channel()
	if err != nil {
		_ = connection.Close()
		return nil, fmt.Errorf("abrir canal RabbitMQ: %w", err)
	}
	consumerChannel, err := connection.Channel()
	if err != nil {
		_ = channel.Close()
		_ = connection.Close()
		return nil, fmt.Errorf("abrir canal consumidor: %w", err)
	}
	closeOnError := func(err error) (*RabbitMQPublisher, error) {
		_ = consumerChannel.Close()
		_ = channel.Close()
		_ = connection.Close()
		return nil, err
	}
	if err := channel.ExchangeDeclare(
		EventsExchange, "topic", true, false, false, false, nil,
	); err != nil {
		return closeOnError(fmt.Errorf("declarar exchange: %w", err))
	}
	if err := channel.ExchangeDeclare(
		DeadLetterExchange, "topic", true, false, false, false, nil,
	); err != nil {
		return closeOnError(fmt.Errorf("declarar dead-letter exchange: %w", err))
	}
	if _, err := channel.QueueDeclare(
		BaixaSolicitadaQueue, true, false, false, false, nil,
	); err != nil {
		return closeOnError(fmt.Errorf("declarar fila: %w", err))
	}
	if err := channel.QueueBind(
		BaixaSolicitadaQueue,
		application.EventTypeBaixaSolicitada,
		EventsExchange,
		false,
		nil,
	); err != nil {
		return closeOnError(fmt.Errorf("vincular fila: %w", err))
	}
	if _, err := consumerChannel.QueueDeclare(
		ResultadoBaixaQueue, true, false, false, false, nil,
	); err != nil {
		return closeOnError(fmt.Errorf("declarar fila de resultados: %w", err))
	}
	if _, err := consumerChannel.QueueDeclare(
		ResultadoBaixaDLQ, true, false, false, false, nil,
	); err != nil {
		return closeOnError(fmt.Errorf("declarar DLQ de resultados: %w", err))
	}
	if _, err := consumerChannel.QueueDeclare(
		ResultadoBaixaRetry, true, false, false, false, amqp.Table{
			"x-dead-letter-exchange":    EventsExchange,
			"x-dead-letter-routing-key": ResultadoRetryRoute,
		},
	); err != nil {
		return closeOnError(fmt.Errorf("declarar fila de retentativa: %w", err))
	}
	if err := consumerChannel.QueueBind(
		ResultadoBaixaDLQ, ResultadoBaixaQueue, DeadLetterExchange, false, nil,
	); err != nil {
		return closeOnError(fmt.Errorf("vincular DLQ de resultados: %w", err))
	}
	for _, eventType := range []string{
		notafiscalApplication.EventTypeBaixaRealizada,
		notafiscalApplication.EventTypeBaixaRejeitada,
		ResultadoRetryRoute,
	} {
		if err := consumerChannel.QueueBind(
			ResultadoBaixaQueue, eventType, EventsExchange, false, nil,
		); err != nil {
			return closeOnError(fmt.Errorf("vincular fila de resultados: %w", err))
		}
	}
	if err := consumerChannel.Qos(1, 0, false); err != nil {
		return closeOnError(fmt.Errorf("configurar prefetch: %w", err))
	}
	if err := channel.Confirm(false); err != nil {
		return closeOnError(fmt.Errorf("habilitar confirmacao de publicacao: %w", err))
	}
	return &RabbitMQPublisher{
		connection:        connection,
		channel:           channel,
		consumerChannel:   consumerChannel,
		confirmations:     channel.NotifyPublish(make(chan amqp.Confirmation, 1)),
		messageTimeout:    messageTimeout,
		messageMaxRetries: messageMaxRetries,
		messageRetryDelay: messageRetryDelay,
	}, nil
}

func (publisher *RabbitMQPublisher) ConsumirResultados(
	ctx context.Context,
	useCase *notafiscalApplication.ProcessarResultadoBaixaUseCase,
) error {
	deliveries, err := publisher.consumerChannel.ConsumeWithContext(
		ctx, ResultadoBaixaQueue, "", false, false, false, false, nil,
	)
	if err != nil {
		return fmt.Errorf("iniciar consumidor de resultados: %w", err)
	}
	for delivery := range deliveries {
		messageContext, cancel := context.WithTimeout(ctx, publisher.messageTimeout)
		var message resultadoBaixaMessage
		if err := json.Unmarshal(delivery.Body, &message); err != nil ||
			message.EventID == uuid.Nil ||
			message.CorrelationID == uuid.Nil ||
			message.NotaFiscalID == uuid.Nil {
			if err := publisher.enviarResultadoParaDLQ(messageContext, delivery); err != nil {
				_ = delivery.Nack(false, true)
				cancel()
				continue
			}
			_ = delivery.Ack(false)
			cancel()
			continue
		}
		_, err := useCase.Execute(messageContext, notafiscalApplication.ResultadoBaixaInput{
			EventID: message.EventID, CorrelationID: message.CorrelationID,
			NotaFiscalID: message.NotaFiscalID, Type: delivery.Type, Motivo: message.Motivo,
		})
		if err != nil {
			if erroResultadoTerminal(err) {
				if err := publisher.enviarResultadoParaDLQ(ctx, delivery); err == nil {
					_ = delivery.Ack(false)
					cancel()
					continue
				}
			}
			if err := publisher.reagendarOuEnviarParaDLQ(ctx, delivery); err != nil {
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

func (publisher *RabbitMQPublisher) reagendarOuEnviarParaDLQ(ctx context.Context, delivery amqp.Delivery) error {
	retry := retryCount(delivery.Headers)
	message := copiarMensagem(delivery)
	if retry >= publisher.messageMaxRetries {
		return publisher.publicarComConfirmacao(ctx, DeadLetterExchange, ResultadoBaixaQueue, message)
	}
	message.Headers["x-retry-count"] = int32(retry + 1)
	message.Expiration = fmt.Sprintf("%d", publisher.messageRetryDelay.Milliseconds())
	return publisher.publicarComConfirmacao(ctx, "", ResultadoBaixaRetry, message)
}

func copiarMensagem(delivery amqp.Delivery) amqp.Publishing {
	headers := amqp.Table{}
	for key, value := range delivery.Headers {
		headers[key] = value
	}
	return amqp.Publishing{Headers: headers, ContentType: delivery.ContentType, DeliveryMode: amqp.Persistent, MessageId: delivery.MessageId, CorrelationId: delivery.CorrelationId, Timestamp: delivery.Timestamp, Type: delivery.Type, Body: delivery.Body}
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

func erroResultadoTerminal(err error) bool {
	return errors.Is(err, notafiscalApplication.ErrTipoResultadoInvalido) ||
		errors.Is(err, notafiscal.ErrNotaNaoEncontrada) ||
		errors.Is(err, notafiscal.ErrNotaNaoEstaProcessando) ||
		errors.Is(err, notafiscal.ErrMotivoRejeicaoObrigatorio)
}

func (publisher *RabbitMQPublisher) enviarResultadoParaDLQ(
	ctx context.Context,
	delivery amqp.Delivery,
) error {
	return publisher.publicarComConfirmacao(
		ctx,
		DeadLetterExchange,
		ResultadoBaixaQueue,
		amqp.Publishing{
			Headers: delivery.Headers, ContentType: delivery.ContentType,
			DeliveryMode: amqp.Persistent, MessageId: delivery.MessageId,
			CorrelationId: delivery.CorrelationId, Timestamp: delivery.Timestamp,
			Type: delivery.Type, Body: delivery.Body,
		},
	)
}

func (publisher *RabbitMQPublisher) Publicar(
	ctx context.Context,
	event application.Event,
) error {
	return publisher.publicarComConfirmacao(
		ctx,
		EventsExchange,
		event.Type,
		amqp.Publishing{
			ContentType: "application/json", DeliveryMode: amqp.Persistent,
			MessageId: event.ID.String(), Timestamp: event.CreatedAt,
			Type: event.Type, Body: event.Payload,
		},
	)
}

func (publisher *RabbitMQPublisher) publicarComConfirmacao(
	ctx context.Context,
	exchange, routingKey string,
	message amqp.Publishing,
) error {
	requestContext, cancel := context.WithTimeout(ctx, publisher.messageTimeout)
	defer cancel()
	publisher.mutex.Lock()
	defer publisher.mutex.Unlock()
	if err := publisher.channel.PublishWithContext(
		requestContext,
		exchange,
		routingKey,
		false,
		false,
		message,
	); err != nil {
		return fmt.Errorf("publicar em %s com routing key %s: %w", exchange, routingKey, err)
	}
	select {
	case confirmation, ok := <-publisher.confirmations:
		if !ok || !confirmation.Ack {
			return ErrPublicacaoNaoConfirmada
		}
		return nil
	case <-requestContext.Done():
		return requestContext.Err()
	}
}

func (publisher *RabbitMQPublisher) Close() error {
	channelErr := publisher.channel.Close()
	consumerErr := publisher.consumerChannel.Close()
	connectionErr := publisher.connection.Close()
	return errors.Join(channelErr, consumerErr, connectionErr)
}
