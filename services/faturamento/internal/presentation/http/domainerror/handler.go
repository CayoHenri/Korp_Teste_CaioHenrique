package domainerror

import (
	"errors"
	application "github.com/caiog/korp-notas-fiscais/services/faturamento/internal/application/nota_fiscal"
	domain "github.com/caiog/korp-notas-fiscais/services/faturamento/internal/domain/nota_fiscal"
	"github.com/caiog/korp-notas-fiscais/services/faturamento/internal/presentation/http/response"
	"github.com/gin-gonic/gin"
	"net/http"
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
	{domain.ErrNotaNaoEncontrada, Mapping{http.StatusNotFound, "NOTA_NAO_ENCONTRADA", domain.ErrNotaNaoEncontrada.Error()}},
	{domain.ErrNotaNaoEstaAberta, Mapping{http.StatusConflict, "NOTA_NAO_ESTA_ABERTA", domain.ErrNotaNaoEstaAberta.Error()}},
	{domain.ErrNotaNaoEstaProcessando, Mapping{http.StatusConflict, "NOTA_NAO_ESTA_PROCESSANDO", domain.ErrNotaNaoEstaProcessando.Error()}},
	{domain.ErrNotaSemItens, Mapping{http.StatusBadRequest, "NOTA_SEM_ITENS", domain.ErrNotaSemItens.Error()}},
	{domain.ErrQuantidadeInvalida, Mapping{http.StatusBadRequest, "QUANTIDADE_INVALIDA", domain.ErrQuantidadeInvalida.Error()}},
	{domain.ErrProdutoInvalido, Mapping{http.StatusBadRequest, "PRODUTO_INVALIDO", domain.ErrProdutoInvalido.Error()}},
	{domain.ErrCodigoObrigatorio, Mapping{http.StatusBadRequest, "CODIGO_PRODUTO_OBRIGATORIO", domain.ErrCodigoObrigatorio.Error()}},
	{domain.ErrDescricaoObrigatoria, Mapping{http.StatusBadRequest, "DESCRICAO_PRODUTO_OBRIGATORIA", domain.ErrDescricaoObrigatoria.Error()}},
	{domain.ErrValorInvalido, Mapping{http.StatusBadRequest, "VALOR_INVALIDO", domain.ErrValorInvalido.Error()}},
	{domain.ErrProdutoInativo, Mapping{http.StatusConflict, "PRODUTO_INATIVO", domain.ErrProdutoInativo.Error()}},
	{domain.ErrNomeClienteObrigatorio, Mapping{http.StatusBadRequest, "NOME_CLIENTE_OBRIGATORIO", domain.ErrNomeClienteObrigatorio.Error()}},
	{domain.ErrEnderecoClienteObrigatorio, Mapping{http.StatusBadRequest, "ENDERECO_CLIENTE_OBRIGATORIO", domain.ErrEnderecoClienteObrigatorio.Error()}},
	{application.ErrProdutoNaoEncontrado, Mapping{http.StatusUnprocessableEntity, "PRODUTO_NAO_ENCONTRADO", application.ErrProdutoNaoEncontrado.Error()}},
	{application.ErrEstoqueIndisponivel, Mapping{http.StatusBadGateway, "ESTOQUE_INDISPONIVEL", application.ErrEstoqueIndisponivel.Error()}},
}

func Map(err error) Mapping {
	for _, entry := range mappings {
		if errors.Is(err, entry.Target) {
			return entry.Mapping
		}
	}
	return Mapping{http.StatusInternalServerError, "ERRO_INTERNO", "erro interno do servidor"}
}
func Respond(c *gin.Context, err error) {
	mapping := Map(err)
	response.Error(c, mapping.Status, mapping.Code, mapping.Message)
}
