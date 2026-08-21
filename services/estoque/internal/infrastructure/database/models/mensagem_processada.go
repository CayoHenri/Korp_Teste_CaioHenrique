package models

import (
	"time"

	"github.com/google/uuid"
)

type MensagemProcessada struct {
	EventID     uuid.UUID `gorm:"column:event_id;type:uuid;primaryKey"`
	ProcessedAt time.Time `gorm:"column:processed_at"`
}

func (MensagemProcessada) TableName() string {
	return "estoque.mensagens_processadas"
}
