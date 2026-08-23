package repository

import (
	"context"
	"time"

	application "github.com/caiog/korp-notas-fiscais/services/faturamento/internal/application/outbox"
	"github.com/caiog/korp-notas-fiscais/services/faturamento/internal/infrastructure/database/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type OutboxRepository struct {
	db *gorm.DB
}

func NewOutboxRepository(db *gorm.DB) *OutboxRepository {
	return &OutboxRepository{db: db}
}

func (repository *OutboxRepository) ListarPendentes(
	ctx context.Context,
	limit int,
) ([]application.Event, error) {
	var records []models.OutboxEvent
	if err := repository.db.WithContext(ctx).
		Where("published_at IS NULL").
		Order("created_at ASC").
		Limit(limit).
		Find(&records).Error; err != nil {
		return nil, err
	}
	events := make([]application.Event, 0, len(records))
	for _, record := range records {
		events = append(events, application.Event{
			ID:        record.ID,
			Type:      record.EventType,
			Payload:   append([]byte(nil), record.Payload...),
			CreatedAt: record.CreatedAt,
		})
	}
	return events, nil
}

func (repository *OutboxRepository) MarcarPublicado(
	ctx context.Context,
	id uuid.UUID,
	publishedAt time.Time,
) error {
	return repository.db.WithContext(ctx).
		Model(&models.OutboxEvent{}).
		Where("id = ? AND published_at IS NULL", id).
		Update("published_at", publishedAt.UTC()).Error
}
