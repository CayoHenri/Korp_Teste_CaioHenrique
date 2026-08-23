package models

import (
	"time"

	"github.com/google/uuid"
)

type MensagemProcessada struct {
	CorrelationID uuid.UUID `gorm:"type:uuid;primaryKey"`
	EventID       uuid.UUID `gorm:"type:uuid;uniqueIndex"`
	ProcessedAt   time.Time
}

func (MensagemProcessada) TableName() string {
	return "faturamento.mensagens_processadas"
}
