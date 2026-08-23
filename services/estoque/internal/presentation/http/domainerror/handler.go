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

var mappings = []struct {
	Target  error
	Mapping Mapping
}{
	{domain.ErrCodigoObrigatorio, Mapping{http.StatusBadRequest, "CODIGO_OBRIGATORIO", domain.ErrCodigoObrigatorio.Error()}},
	{domain.ErrDescricaoObrigatoria, Mapping{http.StatusBadRequest, "DESCRICAO_OBRIGATORIA", domain.ErrDescricaoObrigatoria.Error()}},
	{domain.ErrSaldoInvalido, Mapping{http.StatusBadRequest, "SALDO_INVALIDO", domain.ErrSaldoInvalido.Error()}},
	{domain.ErrValorInvalido, Mapping{http.StatusBadRequest, "VALOR_INVALIDO", domain.ErrValorInvalido.Error()}},
	{domain.ErrCodigoJaExistente, Mapping{http.StatusConflict, "CODIGO_PRODUTO_JA_EXISTENTE", domain.ErrCodigoJaExistente.Error()}},
	{domain.ErrProdutoNaoEncontrado, Mapping{http.StatusNotFound, "PRODUTO_NAO_ENCONTRADO", domain.ErrProdutoNaoEncontrado.Error()}},
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
