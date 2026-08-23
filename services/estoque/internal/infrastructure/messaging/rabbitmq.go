package messaging

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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
	ResultadoBaixaQueue  = "faturamento.baixa.resultado"
)

var ErrPublicacaoNaoConfirmada = errors.New("publicacao nao confirmada pelo RabbitMQ")

type RabbitMQ struct {
	connection       *amqp.Connection
	consumerChannel  *amqp.Channel
	publisherChannel *amqp.Channel
	confirmations    <-chan amqp.Confirmation
	publisherMutex   sync.Mutex
}

func NewRabbitMQ(url string) (*RabbitMQ, error) {
	connection, err := amqp.Dial(url)
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
		connection:       connection,
		consumerChannel:  consumerChannel,
		publisherChannel: publisherChannel,
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
		var message BaixaSolicitadaMessage
		if err := json.Unmarshal(delivery.Body, &message); err != nil {
			if err := client.publicarMensagemInvalida(ctx, delivery); err != nil {
				_ = delivery.Nack(false, true)
				continue
			}
			_ = delivery.Ack(false)
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
		if err := useCase.Execute(ctx, input); err != nil {
			_ = delivery.Nack(false, true)
			continue
		}
		_ = delivery.Ack(false)
	}
	return ctx.Err()
}

func (client *RabbitMQ) publicarMensagemInvalida(
	ctx context.Context,
	delivery amqp.Delivery,
) error {
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

func (client *RabbitMQ) PublicarResultado(
	ctx context.Context,
	result application.ResultadoBaixa,
) error {
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
	client.publisherMutex.Lock()
	defer client.publisherMutex.Unlock()
	if err := client.publisherChannel.PublishWithContext(
		ctx, exchange, routingKey, false, false, message,
	); err != nil {
		return err
	}
	select {
	case confirmation, ok := <-client.confirmations:
		if !ok || !confirmation.Ack {
			return ErrPublicacaoNaoConfirmada
		}
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (client *RabbitMQ) Close() error {
	return errors.Join(
		client.consumerChannel.Close(),
		client.publisherChannel.Close(),
		client.connection.Close(),
	)
}
