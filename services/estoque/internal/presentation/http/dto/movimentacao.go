package dto

import (
	"time"

	domain "github.com/caiog/korp-notas-fiscais/services/estoque/internal/domain/movimentacao"
	"github.com/google/uuid"
)

type MovimentacaoResponse struct {
	ID               uuid.UUID  `json:"id"`
	ProdutoID        uuid.UUID  `json:"produtoId"`
	Tipo             string     `json:"tipo"`
	Quantidade       int        `json:"quantidade"`
	Referencia       *uuid.UUID `json:"referencia,omitempty"`
	DataMovimentacao time.Time  `json:"dataMovimentacao"`
}

func NewMovimentacaoResponse(movimentacao *domain.Movimentacao) MovimentacaoResponse {
	return MovimentacaoResponse{
		ID:               movimentacao.ID(),
		ProdutoID:        movimentacao.ProdutoID(),
		Tipo:             string(movimentacao.Tipo()),
		Quantidade:       movimentacao.Quantidade(),
		Referencia:       movimentacao.Referencia(),
		DataMovimentacao: movimentacao.DataMovimentacao(),
	}
}
