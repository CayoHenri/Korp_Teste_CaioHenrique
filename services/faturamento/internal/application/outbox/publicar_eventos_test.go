package outbox

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

type repositoryStub struct {
	events  []Event
	marked  []uuid.UUID
	listErr error
	markErr error
}

func (repository *repositoryStub) ListarPendentes(context.Context, int) ([]Event, error) {
	return repository.events, repository.listErr
}

func (repository *repositoryStub) MarcarPublicado(
	_ context.Context,
	id uuid.UUID,
	_ time.Time,
) error {
	if repository.markErr != nil {
		return repository.markErr
	}
	repository.marked = append(repository.marked, id)
	return nil
}

type publisherStub struct {
	published []Event
	err       error
}

func (publisher *publisherStub) Publicar(_ context.Context, event Event) error {
	if publisher.err != nil {
		return publisher.err
	}
	publisher.published = append(publisher.published, event)
	return nil
}

func TestPublicarEventosMarcaSomenteAposPublicacaoConfirmada(t *testing.T) {
	event := Event{ID: uuid.New(), Type: EventTypeBaixaSolicitada}
	repository := &repositoryStub{events: []Event{event}}
	publisher := &publisherStub{}

	published, err := NewPublicarEventosUseCase(repository, publisher, 100).
		Execute(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if published != 1 || len(publisher.published) != 1 || len(repository.marked) != 1 {
		t.Fatal("evento deveria ser publicado e marcado")
	}
}

func TestPublicarEventosNaoMarcaQuandoPublisherFalha(t *testing.T) {
	publishErr := errors.New("rabbit indisponivel")
	repository := &repositoryStub{events: []Event{{ID: uuid.New()}}}
	publisher := &publisherStub{err: publishErr}

	published, err := NewPublicarEventosUseCase(repository, publisher, 100).
		Execute(context.Background())
	if !errors.Is(err, publishErr) || published != 0 {
		t.Fatalf("resultado inesperado: published=%d err=%v", published, err)
	}
	if len(repository.marked) != 0 {
		t.Fatal("evento com falha de publicacao nao pode ser marcado")
	}
}
