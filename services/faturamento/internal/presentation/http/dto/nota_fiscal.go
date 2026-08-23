package dto

import (
	"time"

	domain "github.com/caiog/korp-notas-fiscais/services/faturamento/internal/domain/nota_fiscal"
	"github.com/google/uuid"
)

type CriarNotaFiscalItemRequest struct {
	CodigoProduto string `json:"codigoProduto" binding:"required" example:"SKU-001"`
	Quantidade    int    `json:"quantidade" binding:"required" example:"2"`
}

type CriarNotaFiscalRequest struct {
	NomeCliente     string                       `json:"nomeCliente" binding:"required" example:"MARIA DA SILVA"`
	EnderecoCliente string                       `json:"enderecoCliente" binding:"required" example:"RUA DAS FLORES, 100 - CURITIBA/PR"`
	Itens           []CriarNotaFiscalItemRequest `json:"itens" binding:"required"`
}

type ItemNotaFiscalResponse struct {
	ID               uuid.UUID `json:"id" example:"550e8400-e29b-41d4-a716-446655440001"`
	ProdutoID        uuid.UUID `json:"produtoId" example:"550e8400-e29b-41d4-a716-446655440002"`
	CodigoProduto    string    `json:"codigoProduto" example:"SKU-001"`
	DescricaoProduto string    `json:"descricaoProduto" example:"TECLADO MECANICO"`
	Quantidade       int       `json:"quantidade" example:"2"`
	Valor            float64   `json:"valor" example:"159.90"`
	ValorTotal       float64   `json:"valorTotal" example:"319.80"`
}

type NotaFiscalResponse struct {
	ID              uuid.UUID                `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Numero          int64                    `json:"numero" example:"1001"`
	Status          string                   `json:"status" enums:"ABERTA,PROCESSANDO,FECHADA" example:"ABERTA"`
	NomeCliente     string                   `json:"nomeCliente" example:"MARIA DA SILVA"`
	EnderecoCliente string                   `json:"enderecoCliente" example:"RUA DAS FLORES, 100 - CURITIBA/PR"`
	QuantidadeTotal int                      `json:"quantidadeTotal" example:"2"`
	ValorTotal      float64                  `json:"valorTotal" example:"319.80"`
	Itens           []ItemNotaFiscalResponse `json:"itens"`
	DataCadastro    time.Time                `json:"dataCadastro" example:"2026-08-23T12:00:00Z"`
	DataAtualizacao time.Time                `json:"dataAtualizacao" example:"2026-08-23T12:00:00Z"`
	DataFechamento  *time.Time               `json:"dataFechamento,omitempty" example:"2026-08-23T12:05:00Z"`
}

func NewNotaFiscalResponse(nota *domain.NotaFiscal) NotaFiscalResponse {
	itens := nota.Itens()
	result := NotaFiscalResponse{
		ID:              nota.ID(),
		Numero:          nota.Numero(),
		Status:          string(nota.Status()),
		NomeCliente:     nota.NomeCliente(),
		EnderecoCliente: nota.EnderecoCliente(),
		QuantidadeTotal: nota.QuantidadeTotal(),
		ValorTotal:      nota.ValorTotal(),
		Itens:           make([]ItemNotaFiscalResponse, 0, len(itens)),
		DataCadastro:    nota.DataCadastro(),
		DataAtualizacao: nota.DataAtualizacao(),
		DataFechamento:  nota.DataFechamento(),
	}
	for _, item := range itens {
		result.Itens = append(result.Itens, ItemNotaFiscalResponse{
			ID:               item.ID(),
			ProdutoID:        item.ProdutoID(),
			CodigoProduto:    item.CodigoProduto(),
			DescricaoProduto: item.DescricaoProduto(),
			Quantidade:       item.Quantidade(),
			Valor:            item.Valor(),
			ValorTotal:       item.ValorTotal(),
		})
	}
	return result
}
