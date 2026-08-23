package dto

import (
	domain "github.com/caiog/korp-notas-fiscais/services/faturamento/internal/domain/nota_fiscal"
	"github.com/google/uuid"
	"time"
)

type CriarNotaFiscalItemRequest struct {
	CodigoProduto string `json:"codigoProduto" binding:"required"`
	Quantidade    int    `json:"quantidade" binding:"required"`
}
type CriarNotaFiscalRequest struct {
	NomeCliente     string                       `json:"nomeCliente" binding:"required"`
	EnderecoCliente string                       `json:"enderecoCliente" binding:"required"`
	Itens           []CriarNotaFiscalItemRequest `json:"itens" binding:"required"`
}
type ItemNotaFiscalResponse struct {
	ID               uuid.UUID `json:"id"`
	ProdutoID        uuid.UUID `json:"produtoId"`
	CodigoProduto    string    `json:"codigoProduto"`
	DescricaoProduto string    `json:"descricaoProduto"`
	Quantidade       int       `json:"quantidade"`
	Valor            float64   `json:"valor"`
	ValorTotal       float64   `json:"valorTotal"`
}
type NotaFiscalResponse struct {
	ID              uuid.UUID                `json:"id"`
	Numero          int64                    `json:"numero"`
	Status          string                   `json:"status"`
	NomeCliente     string                   `json:"nomeCliente"`
	EnderecoCliente string                   `json:"enderecoCliente"`
	QuantidadeTotal int                      `json:"quantidadeTotal"`
	ValorTotal      float64                  `json:"valorTotal"`
	Itens           []ItemNotaFiscalResponse `json:"itens"`
	DataCadastro    time.Time                `json:"dataCadastro"`
	DataAtualizacao time.Time                `json:"dataAtualizacao"`
	DataFechamento  *time.Time               `json:"dataFechamento,omitempty"`
}

func NewNotaFiscalResponse(nota *domain.NotaFiscal) NotaFiscalResponse {
	itens := nota.Itens()
	result := NotaFiscalResponse{ID: nota.ID(), Numero: nota.Numero(), Status: string(nota.Status()), NomeCliente: nota.NomeCliente(), EnderecoCliente: nota.EnderecoCliente(), QuantidadeTotal: nota.QuantidadeTotal(), ValorTotal: nota.ValorTotal(), Itens: make([]ItemNotaFiscalResponse, 0, len(itens)), DataCadastro: nota.DataCadastro(), DataAtualizacao: nota.DataAtualizacao(), DataFechamento: nota.DataFechamento()}
	for _, item := range itens {
		result.Itens = append(result.Itens, ItemNotaFiscalResponse{ID: item.ID(), ProdutoID: item.ProdutoID(), CodigoProduto: item.CodigoProduto(), DescricaoProduto: item.DescricaoProduto(), Quantidade: item.Quantidade(), Valor: item.Valor(), ValorTotal: item.ValorTotal()})
	}
	return result
}
