package dto

import (
	"time"

	domain "github.com/caiog/korp-notas-fiscais/services/estoque/internal/domain/produto"
	"github.com/google/uuid"
)

type CriarProdutoRequest struct {
	Codigo    string  `json:"codigo" example:"SKU-001"`
	Descricao string  `json:"descricao" example:"TECLADO MECANICO"`
	Saldo     int     `json:"saldo" example:"10"`
	Valor     float64 `json:"valor" binding:"required" example:"15990"`
}

type AtualizarProdutoRequest struct {
	Descricao string   `json:"descricao" binding:"required" example:"TECLADO MECANICO RGB"`
	Saldo     *int     `json:"saldo" binding:"required" example:"20"`
	Valor     *float64 `json:"valor" binding:"required" example:"15990"`
}

type ProdutoResponse struct {
	ID              uuid.UUID `json:"id"`
	Codigo          string    `json:"codigo"`
	Descricao       string    `json:"descricao"`
	Saldo           int       `json:"saldo"`
	Valor           float64   `json:"valor"`
	Ativo           bool      `json:"ativo"`
	DataCadastro    time.Time `json:"dataCadastro"`
	DataAtualizacao time.Time `json:"dataAtualizacao"`
}

func NewProdutoResponse(produto *domain.Produto) ProdutoResponse {
	return ProdutoResponse{
		ID:              produto.ID(),
		Codigo:          produto.Codigo(),
		Descricao:       produto.Descricao(),
		Saldo:           produto.Saldo(),
		Valor:           produto.Valor(),
		Ativo:           produto.Ativo(),
		DataCadastro:    produto.DataCadastro(),
		DataAtualizacao: produto.DataAtualizacao(),
	}
}
