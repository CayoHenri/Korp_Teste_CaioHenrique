package http

import (
	"net/http"

	estoqueApplication "github.com/caiog/korp-notas-fiscais/services/estoque/internal/application/estoque"
	application "github.com/caiog/korp-notas-fiscais/services/estoque/internal/application/produto"
	domain "github.com/caiog/korp-notas-fiscais/services/estoque/internal/domain/produto"
	"github.com/caiog/korp-notas-fiscais/services/estoque/internal/presentation/http/domainerror"
	"github.com/caiog/korp-notas-fiscais/services/estoque/internal/presentation/http/dto"
	"github.com/caiog/korp-notas-fiscais/services/estoque/internal/presentation/http/response"
	sharedquery "github.com/caiog/korp-notas-fiscais/services/estoque/internal/shared/query"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ProdutoHandler struct {
	criarProduto           *application.CriarProdutoUseCase
	listarProdutos         *application.ListarProdutosUseCase
	buscarProdutoPorID     *application.BuscarProdutoPorIDUseCase
	buscarProdutoPorCodigo *application.BuscarProdutoPorCodigoUseCase
	ativarProduto          *application.AtivarProdutoUseCase
	inativarProduto        *application.InativarProdutoUseCase
	atualizarProduto       *application.AtualizarProdutoUseCase
	listarMovimentacoes    *estoqueApplication.ListarMovimentacoesUseCase
}

func NewProdutoHandler(
	criarProduto *application.CriarProdutoUseCase,
	listarProdutos *application.ListarProdutosUseCase,
	buscarProdutoPorID *application.BuscarProdutoPorIDUseCase,
	buscarProdutoPorCodigo *application.BuscarProdutoPorCodigoUseCase,
	ativarProduto *application.AtivarProdutoUseCase,
	inativarProduto *application.InativarProdutoUseCase,
	atualizarProduto *application.AtualizarProdutoUseCase,
	listarMovimentacoes *estoqueApplication.ListarMovimentacoesUseCase,
) *ProdutoHandler {
	return &ProdutoHandler{
		criarProduto:           criarProduto,
		listarProdutos:         listarProdutos,
		buscarProdutoPorID:     buscarProdutoPorID,
		buscarProdutoPorCodigo: buscarProdutoPorCodigo,
		ativarProduto:          ativarProduto,
		inativarProduto:        inativarProduto,
		atualizarProduto:       atualizarProduto,
		listarMovimentacoes:    listarMovimentacoes,
	}
}

func (handler *ProdutoHandler) RegisterRoutes(router *gin.Engine) {
	group := router.Group("/produtos")
	group.POST("", handler.criar)
	group.GET("", handler.listar)
	group.GET("/:id", handler.buscarPorID)
	group.GET("/codigo/:codigo", handler.buscarPorCodigo)
	group.PATCH("/:id/ativar", handler.ativar)
	group.PATCH("/:id/inativar", handler.inativar)
	group.PUT("/:id", handler.atualizar)
	group.GET("/:id/movimentacoes", handler.listarHistorico)
}

// listarHistorico godoc
// @Summary Lista o historico de movimentacoes de um produto
// @Tags Produtos
// @Produce json
// @Param id path string true "UUID do produto"
// @Success 200 {object} response.SuccessResponse{data=[]dto.MovimentacaoResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /produtos/{id}/movimentacoes [get]
func (handler *ProdutoHandler) listarHistorico(c *gin.Context) {
	id, ok := produtoID(c)
	if !ok {
		return
	}
	movimentacoes, err := handler.listarMovimentacoes.Execute(c.Request.Context(), id)
	if err != nil {
		domainerror.Respond(c, err)
		return
	}
	result := make([]dto.MovimentacaoResponse, 0, len(movimentacoes))
	for index := range movimentacoes {
		result = append(result, dto.NewMovimentacaoResponse(&movimentacoes[index]))
	}
	response.OK(c, result)
}

// atualizar godoc
// @Summary Atualiza descricao, valor e saldo de um produto
// @Tags Produtos
// @Accept json
// @Produce json
// @Param id path string true "UUID do produto"
// @Param produto body dto.AtualizarProdutoRequest true "Campos atualizaveis"
// @Success 200 {object} response.SuccessResponse{data=dto.ProdutoResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /produtos/{id} [put]
func (handler *ProdutoHandler) atualizar(c *gin.Context) {
	id, ok := produtoID(c)
	if !ok {
		return
	}

	var request dto.AtualizarProdutoRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.Error(c, http.StatusBadRequest, "REQUISICAO_INVALIDA", "descricao e saldo sao obrigatorios")
		return
	}

	produto, err := handler.atualizarProduto.Execute(
		c.Request.Context(), application.AtualizarProdutoInput{
			ID:        id,
			Descricao: request.Descricao,
			Saldo:     *request.Saldo,
			Valor:     *request.Valor,
		},
	)
	if err != nil {
		domainerror.Respond(c, err)
		return
	}
	response.OK(c, dto.NewProdutoResponse(produto))
}

// ativar godoc
// @Summary Ativa um produto
// @Tags Produtos
// @Produce json
// @Param id path string true "UUID do produto"
// @Success 200 {object} response.SuccessResponse{data=dto.ProdutoResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /produtos/{id}/ativar [patch]
func (handler *ProdutoHandler) ativar(c *gin.Context) {
	id, ok := produtoID(c)
	if !ok {
		return
	}
	produto, err := handler.ativarProduto.Execute(c.Request.Context(), id)
	if err != nil {
		domainerror.Respond(c, err)
		return
	}
	response.OK(c, dto.NewProdutoResponse(produto))
}

// inativar godoc
// @Summary Inativa um produto sem exclui-lo
// @Tags Produtos
// @Produce json
// @Param id path string true "UUID do produto"
// @Success 200 {object} response.SuccessResponse{data=dto.ProdutoResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /produtos/{id}/inativar [patch]
func (handler *ProdutoHandler) inativar(c *gin.Context) {
	id, ok := produtoID(c)
	if !ok {
		return
	}
	produto, err := handler.inativarProduto.Execute(c.Request.Context(), id)
	if err != nil {
		domainerror.Respond(c, err)
		return
	}
	response.OK(c, dto.NewProdutoResponse(produto))
}

func produtoID(c *gin.Context) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "ID_INVALIDO", "id do produto deve ser um UUID valido")
		return uuid.Nil, false
	}
	return id, true
}

// criar godoc
// @Summary Cadastra um produto
// @Tags Produtos
// @Accept json
// @Produce json
// @Param produto body dto.CriarProdutoRequest true "Dados do produto"
// @Success 201 {object} response.SuccessResponse{data=dto.ProdutoResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /produtos [post]
func (handler *ProdutoHandler) criar(c *gin.Context) {
	var request dto.CriarProdutoRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.Error(c, http.StatusBadRequest, "REQUISICAO_INVALIDA", "corpo da requisicao invalido")
		return
	}

	produto, err := handler.criarProduto.Execute(c.Request.Context(), application.CriarProdutoInput{
		Codigo:    request.Codigo,
		Descricao: request.Descricao,
		Saldo:     request.Saldo,
		Valor:     request.Valor,
	})
	if err != nil {
		domainerror.Respond(c, err)
		return
	}

	response.Created(c, dto.NewProdutoResponse(produto))
}

// listar godoc
// @Summary Lista produtos
// @Tags Produtos
// @Produce json
// @Param pagina query int false "Pagina" default(1)
// @Param tamanhoPagina query int false "Itens por pagina (maximo 100)" default(20)
// @Param codigo query string false "Trecho do codigo"
// @Param descricao query string false "Trecho da descricao"
// @Param ativo query bool false "Status ativo"
// @Success 200 {object} response.SuccessResponse{data=dto.ProdutosPaginadosResponse}
// @Failure 500 {object} response.ErrorResponse
// @Router /produtos [get]
func (handler *ProdutoHandler) listar(c *gin.Context) {
	var request dto.ListarProdutosQuery
	if err := c.ShouldBindQuery(&request); err != nil {
		response.Error(c, http.StatusBadRequest, "FILTROS_INVALIDOS", "filtros da requisicao invalidos")
		return
	}
	pagina, err := handler.listarProdutos.Execute(
		c.Request.Context(),
		sharedquery.Criteria[domain.ListFilters]{
			Filters: domain.ListFilters{
				Codigo:    request.Codigo,
				Descricao: request.Descricao,
				Ativo:     request.Ativo,
			},
			Pagination: sharedquery.Pagination{
				Page:     request.Pagina,
				PageSize: request.TamanhoPagina,
			},
		},
	)
	if err != nil {
		domainerror.Respond(c, err)
		return
	}

	result := dto.ProdutosPaginadosResponse{
		Itens:         make([]dto.ProdutoResponse, 0, len(pagina.Items)),
		Total:         pagina.Total,
		Pagina:        pagina.Page,
		TamanhoPagina: pagina.PageSize,
		TotalPaginas:  pagina.TotalPages,
	}
	for index := range pagina.Items {
		result.Itens = append(result.Itens, dto.NewProdutoResponse(&pagina.Items[index]))
	}
	response.OK(c, result)
}

// buscarPorID godoc
// @Summary Consulta um produto por ID
// @Tags Produtos
// @Produce json
// @Param id path string true "UUID do produto"
// @Success 200 {object} response.SuccessResponse{data=dto.ProdutoResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /produtos/{id} [get]
func (handler *ProdutoHandler) buscarPorID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "ID_INVALIDO", "id do produto deve ser um UUID valido")
		return
	}

	produto, err := handler.buscarProdutoPorID.Execute(c.Request.Context(), id)
	if err != nil {
		domainerror.Respond(c, err)
		return
	}
	response.OK(c, dto.NewProdutoResponse(produto))
}

// buscarPorCodigo godoc
// @Summary Consulta um produto por codigo
// @Tags Produtos
// @Produce json
// @Param codigo path string true "Codigo do produto"
// @Success 200 {object} response.SuccessResponse{data=dto.ProdutoResponse}
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /produtos/codigo/{codigo} [get]
func (handler *ProdutoHandler) buscarPorCodigo(c *gin.Context) {
	produto, err := handler.buscarProdutoPorCodigo.Execute(c.Request.Context(), c.Param("codigo"))
	if err != nil {
		domainerror.Respond(c, err)
		return
	}
	response.OK(c, dto.NewProdutoResponse(produto))
}
