package models

import (
	"time"

	domain "github.com/caiog/korp-notas-fiscais/services/estoque/internal/domain/movimentacao"
	"github.com/google/uuid"
)

type MovimentacaoEstoque struct {
	ID               uuid.UUID  `gorm:"column:id;type:uuid;primaryKey"`
	ProdutoID        uuid.UUID  `gorm:"column:produto_id;type:uuid"`
	Tipo             string     `gorm:"column:tipo"`
	Quantidade       int        `gorm:"column:quantidade"`
	Referencia       *uuid.UUID `gorm:"column:referencia;type:uuid"`
	DataMovimentacao time.Time  `gorm:"column:data_movimentacao"`
}

func (MovimentacaoEstoque) TableName() string {
	return "estoque.movimentacoes_estoque"
}

func NewMovimentacaoEstoqueModel(movimentacao *domain.Movimentacao) MovimentacaoEstoque {
	return MovimentacaoEstoque{
		ID:               movimentacao.ID(),
		ProdutoID:        movimentacao.ProdutoID(),
		Tipo:             string(movimentacao.Tipo()),
		Quantidade:       movimentacao.Quantidade(),
		Referencia:       movimentacao.Referencia(),
		DataMovimentacao: movimentacao.DataMovimentacao(),
	}
}

func (model MovimentacaoEstoque) ToDomain() (*domain.Movimentacao, error) {
	return domain.NewMovimentacaoWithState(
		model.ID,
		model.ProdutoID,
		domain.Tipo(model.Tipo),
		model.Quantidade,
		model.Referencia,
		model.DataMovimentacao,
	)
}
