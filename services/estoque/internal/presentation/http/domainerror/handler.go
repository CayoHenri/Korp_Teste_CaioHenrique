package domainerror

import (
	"errors"
	"net/http"

	domain "github.com/caiog/korp-notas-fiscais/services/estoque/internal/domain/produto"
	"github.com/caiog/korp-notas-fiscais/services/estoque/internal/presentation/http/response"
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
	newMappingEntry(domain.ErrCodigoObrigatorio, http.StatusBadRequest, "CODIGO_OBRIGATORIO"),
	newMappingEntry(
		domain.ErrDescricaoObrigatoria,
		http.StatusBadRequest,
		"DESCRICAO_OBRIGATORIA",
	),
	newMappingEntry(domain.ErrSaldoInvalido, http.StatusBadRequest, "SALDO_INVALIDO"),
	newMappingEntry(domain.ErrValorInvalido, http.StatusBadRequest, "VALOR_INVALIDO"),
	newMappingEntry(
		domain.ErrCodigoJaExistente,
		http.StatusConflict,
		"CODIGO_PRODUTO_JA_EXISTENTE",
	),
	newMappingEntry(
		domain.ErrProdutoNaoEncontrado,
		http.StatusNotFound,
		"PRODUTO_NAO_ENCONTRADO",
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
