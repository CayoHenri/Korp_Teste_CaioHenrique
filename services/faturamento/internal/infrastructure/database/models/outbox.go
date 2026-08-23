package models

import (
	"github.com/google/uuid"
	"time"
)

type OutboxEvent struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey"`
	EventType   string
	AggregateID uuid.UUID `gorm:"type:uuid"`
	Payload     []byte    `gorm:"type:jsonb"`
	CreatedAt   time.Time
	PublishedAt *time.Time
}

func (OutboxEvent) TableName() string {
	return "faturamento.outbox_events"
}
