package outbox

import (
	"context"
	"time"

	"github.com/google/uuid"
)

const EventTypeBaixaSolicitada = "estoque.baixa.solicitada"

type Event struct {
	ID        uuid.UUID
	Type      string
	Payload   []byte
	CreatedAt time.Time
}

type Repository interface {
	ListarPendentes(context.Context, int) ([]Event, error)
	MarcarPublicado(context.Context, uuid.UUID, time.Time) error
}

type Publisher interface {
	Publicar(context.Context, Event) error
}

type PublicarEventosUseCase struct {
	repository Repository
	publisher  Publisher
	batchSize  int
}

func NewPublicarEventosUseCase(
	repository Repository,
	publisher Publisher,
	batchSize int,
) *PublicarEventosUseCase {
	return &PublicarEventosUseCase{
		repository: repository,
		publisher:  publisher,
		batchSize:  batchSize,
	}
}

func (useCase *PublicarEventosUseCase) Execute(ctx context.Context) (int, error) {
	events, err := useCase.repository.ListarPendentes(ctx, useCase.batchSize)
	if err != nil {
		return 0, err
	}
	published := 0
	for _, event := range events {
		if err := useCase.publisher.Publicar(ctx, event); err != nil {
			return published, err
		}
		if err := useCase.repository.MarcarPublicado(ctx, event.ID, time.Now().UTC()); err != nil {
			return published, err
		}
		published++
	}
	return published, nil
}
