package models

import (
	"time"

	domain "github.com/caiog/korp-notas-fiscais/services/estoque/internal/domain/produto"
	"github.com/google/uuid"
)

type Produto struct {
	ID              uuid.UUID `gorm:"column:id;type:uuid;primaryKey"`
	Codigo          string    `gorm:"column:codigo"`
	Descricao       string    `gorm:"column:descricao"`
	Saldo           int       `gorm:"column:saldo"`
	DataCadastro    time.Time `gorm:"column:data_cadastro"`
	DataAtualizacao time.Time `gorm:"column:data_atualizacao"`
}

func (Produto) TableName() string {
	return "estoque.produtos"
}

func NewProdutoModel(produto *domain.Produto) Produto {
	return Produto{
		ID:              produto.ID(),
		Codigo:          produto.Codigo(),
		Descricao:       produto.Descricao(),
		Saldo:           produto.Saldo(),
		DataCadastro:    produto.DataCadastro(),
		DataAtualizacao: produto.DataAtualizacao(),
	}
}

func (model Produto) ToDomain() (*domain.Produto, error) {
	return domain.NewProdutoWithState(
		model.ID,
		model.Codigo,
		model.Descricao,
		model.Saldo,
		model.DataCadastro,
		model.DataAtualizacao,
	)
}
