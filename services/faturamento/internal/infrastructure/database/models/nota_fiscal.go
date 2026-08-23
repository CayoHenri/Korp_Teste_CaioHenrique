package models

import (
	notafiscal "github.com/caiog/korp-notas-fiscais/services/faturamento/internal/domain/nota_fiscal"
	"github.com/google/uuid"
	"time"
)

type NotaFiscal struct {
	ID              uuid.UUID `gorm:"type:uuid;primaryKey"`
	Numero          int64
	Status          string `gorm:"type:faturamento.nota_fiscal_status"`
	NomeCliente     string
	EnderecoCliente string
	Itens           []ItemNotaFiscal `gorm:"foreignKey:NotaFiscalID"`
	DataCadastro    time.Time
	DataAtualizacao time.Time
	DataFechamento  *time.Time
}

func (NotaFiscal) TableName() string {
	return "faturamento.notas_fiscais"
}

func NewNotaFiscalModel(nota *notafiscal.NotaFiscal) NotaFiscal {
	itens := nota.Itens()
	models := make([]ItemNotaFiscal, 0, len(itens))
	for _, item := range itens {
		models = append(models, NewItemNotaFiscalModel(nota.ID(), &item))
	}
	return NotaFiscal{
		ID:              nota.ID(),
		Numero:          nota.Numero(),
		Status:          string(nota.Status()),
		NomeCliente:     nota.NomeCliente(),
		EnderecoCliente: nota.EnderecoCliente(),
		Itens:           models,
		DataCadastro:    nota.DataCadastro(),
		DataAtualizacao: nota.DataAtualizacao(),
		DataFechamento:  nota.DataFechamento(),
	}
}

func (model *NotaFiscal) ToDomain() (*notafiscal.NotaFiscal, error) {
	itens := make([]notafiscal.ItemNotaFiscal, 0, len(model.Itens))
	for _, itemModel := range model.Itens {
		item, err := itemModel.ToDomain()
		if err != nil {
			return nil, err
		}
		itens = append(itens, *item)
	}
	return notafiscal.NewNotaFiscalWithState(
		model.ID,
		model.Numero,
		notafiscal.Status(model.Status),
		model.NomeCliente,
		model.EnderecoCliente,
		itens,
		model.DataCadastro,
		model.DataAtualizacao,
		model.DataFechamento,
	)
}

type ItemNotaFiscal struct {
	ID               uuid.UUID `gorm:"type:uuid;primaryKey"`
	NotaFiscalID     uuid.UUID `gorm:"type:uuid"`
	ProdutoID        uuid.UUID `gorm:"type:uuid"`
	CodigoProduto    string
	DescricaoProduto string
	Quantidade       int
	Valor            float64
}

func (ItemNotaFiscal) TableName() string {
	return "faturamento.itens_nota_fiscal"
}

func NewItemNotaFiscalModel(notaID uuid.UUID, item *notafiscal.ItemNotaFiscal) ItemNotaFiscal {
	return ItemNotaFiscal{
		ID:               item.ID(),
		NotaFiscalID:     notaID,
		ProdutoID:        item.ProdutoID(),
		CodigoProduto:    item.CodigoProduto(),
		DescricaoProduto: item.DescricaoProduto(),
		Quantidade:       item.Quantidade(),
		Valor:            item.Valor(),
	}
}
func (model *ItemNotaFiscal) ToDomain() (*notafiscal.ItemNotaFiscal, error) {
	return notafiscal.NewItemNotaFiscalWithState(
		model.ID,
		model.ProdutoID,
		model.CodigoProduto,
		model.DescricaoProduto,
		model.Quantidade,
		model.Valor,
	)
}
