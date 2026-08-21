package dto

import (
	"time"

	domain "github.com/caiog/korp-notas-fiscais/services/estoque/internal/domain/produto"
	"github.com/google/uuid"
)

type CriarProdutoRequest struct {
	Codigo    string `json:"codigo" example:"SKU-001"`
	Descricao string `json:"descricao" example:"TECLADO MECANICO"`
	Saldo     int    `json:"saldo" example:"10"`
}

type ProdutoResponse struct {
	ID              uuid.UUID `json:"id"`
	Codigo          string    `json:"codigo"`
	Descricao       string    `json:"descricao"`
	Saldo           int       `json:"saldo"`
	DataCadastro    time.Time `json:"dataCadastro"`
	DataAtualizacao time.Time `json:"dataAtualizacao"`
}

func NewProdutoResponse(produto *domain.Produto) ProdutoResponse {
	return ProdutoResponse{
		ID:              produto.ID(),
		Codigo:          produto.Codigo(),
		Descricao:       produto.Descricao(),
		Saldo:           produto.Saldo(),
		DataCadastro:    produto.DataCadastro(),
		DataAtualizacao: produto.DataAtualizacao(),
	}
}
