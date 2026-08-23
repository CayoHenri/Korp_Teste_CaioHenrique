package domainerror

import (
	"errors"
	"net/http"

	application "github.com/caiog/korp-notas-fiscais/services/faturamento/internal/application/nota_fiscal"
	domain "github.com/caiog/korp-notas-fiscais/services/faturamento/internal/domain/nota_fiscal"
	"github.com/caiog/korp-notas-fiscais/services/faturamento/internal/presentation/http/response"
	"github.com/gin-gonic/gin"
)

type Mapping struct {
	Status  int
	Code    string
	Message string
}

type mappingEntry struct {
	Target  error
	Mapping Mapping
}

var mappings = []mappingEntry{
	newMappingEntry(domain.ErrNotaNaoEncontrada, http.StatusNotFound, "NOTA_NAO_ENCONTRADA"),
	newMappingEntry(domain.ErrNotaNaoEstaAberta, http.StatusConflict, "NOTA_NAO_ESTA_ABERTA"),
	newMappingEntry(
		domain.ErrNotaNaoEstaProcessando,
		http.StatusConflict,
		"NOTA_NAO_ESTA_PROCESSANDO",
	),
	newMappingEntry(domain.ErrNotaSemItens, http.StatusBadRequest, "NOTA_SEM_ITENS"),
	newMappingEntry(domain.ErrQuantidadeInvalida, http.StatusBadRequest, "QUANTIDADE_INVALIDA"),
	newMappingEntry(domain.ErrProdutoInvalido, http.StatusBadRequest, "PRODUTO_INVALIDO"),
	newMappingEntry(domain.ErrCodigoObrigatorio, http.StatusBadRequest, "CODIGO_PRODUTO_OBRIGATORIO"),
	newMappingEntry(
		domain.ErrDescricaoObrigatoria,
		http.StatusBadRequest,
		"DESCRICAO_PRODUTO_OBRIGATORIA",
	),
	newMappingEntry(domain.ErrValorInvalido, http.StatusBadRequest, "VALOR_INVALIDO"),
	newMappingEntry(domain.ErrProdutoInativo, http.StatusConflict, "PRODUTO_INATIVO"),
	newMappingEntry(
		domain.ErrNomeClienteObrigatorio,
		http.StatusBadRequest,
		"NOME_CLIENTE_OBRIGATORIO",
	),
	newMappingEntry(
		domain.ErrEnderecoClienteObrigatorio,
		http.StatusBadRequest,
		"ENDERECO_CLIENTE_OBRIGATORIO",
	),
	newMappingEntry(
		application.ErrProdutoNaoEncontrado,
		http.StatusUnprocessableEntity,
		"PRODUTO_NAO_ENCONTRADO",
	),
	newMappingEntry(
		application.ErrEstoqueIndisponivel,
		http.StatusBadGateway,
		"ESTOQUE_INDISPONIVEL",
	),
}

func newMappingEntry(target error, status int, code string) mappingEntry {
	return mappingEntry{
		Target: target,
		Mapping: Mapping{
			Status:  status,
			Code:    code,
			Message: target.Error(),
		},
	}
}

func Map(err error) Mapping {
	for _, entry := range mappings {
		if errors.Is(err, entry.Target) {
			return entry.Mapping
		}
	}
	return Mapping{
		Status:  http.StatusInternalServerError,
		Code:    "ERRO_INTERNO",
		Message: "erro interno do servidor",
	}
}

func Respond(c *gin.Context, err error) {
	mapping := Map(err)
	response.Error(c, mapping.Status, mapping.Code, mapping.Message)
}
