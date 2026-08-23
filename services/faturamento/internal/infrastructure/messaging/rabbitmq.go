package messaging

import (
	"context"
	"errors"
	"fmt"
	"sync"

	application "github.com/caiog/korp-notas-fiscais/services/faturamento/internal/application/outbox"
	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	EventsExchange       = "korp.events"
	BaixaSolicitadaQueue = "estoque.baixa.solicitada"
)

var ErrPublicacaoNaoConfirmada = errors.New("publicacao nao confirmada pelo RabbitMQ")

type RabbitMQPublisher struct {
	connection    *amqp.Connection
	channel       *amqp.Channel
	confirmations <-chan amqp.Confirmation
	mutex         sync.Mutex
}

func NewRabbitMQPublisher(url string) (*RabbitMQPublisher, error) {
	connection, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("conectar ao RabbitMQ: %w", err)
	}
	channel, err := connection.Channel()
	if err != nil {
		_ = connection.Close()
		return nil, fmt.Errorf("abrir canal RabbitMQ: %w", err)
	}
	closeOnError := func(err error) (*RabbitMQPublisher, error) {
		_ = channel.Close()
		_ = connection.Close()
		return nil, err
	}
	if err := channel.ExchangeDeclare(
		EventsExchange, "topic", true, false, false, false, nil,
	); err != nil {
		return closeOnError(fmt.Errorf("declarar exchange: %w", err))
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
	if err := channel.Confirm(false); err != nil {
		return closeOnError(fmt.Errorf("habilitar confirmacao de publicacao: %w", err))
	}
	return &RabbitMQPublisher{
		connection:    connection,
		channel:       channel,
		confirmations: channel.NotifyPublish(make(chan amqp.Confirmation, 1)),
	}, nil
}

func (publisher *RabbitMQPublisher) Publicar(
	ctx context.Context,
	event application.Event,
) error {
	publisher.mutex.Lock()
	defer publisher.mutex.Unlock()
	if err := publisher.channel.PublishWithContext(
		ctx,
		EventsExchange,
		event.Type,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			MessageId:    event.ID.String(),
			Timestamp:    event.CreatedAt,
			Type:         event.Type,
			Body:         event.Payload,
		},
	); err != nil {
		return fmt.Errorf("publicar evento %s: %w", event.ID, err)
	}
	select {
	case confirmation, ok := <-publisher.confirmations:
		if !ok || !confirmation.Ack {
			return ErrPublicacaoNaoConfirmada
		}
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (publisher *RabbitMQPublisher) Close() error {
	channelErr := publisher.channel.Close()
	connectionErr := publisher.connection.Close()
	return errors.Join(channelErr, connectionErr)
}
