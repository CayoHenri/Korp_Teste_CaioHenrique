package http

import (
	"errors"
	"net/http"
	"time"

	application "github.com/caiog/korp-notas-fiscais/services/estoque/internal/application/produto"
	domain "github.com/caiog/korp-notas-fiscais/services/estoque/internal/domain/produto"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type criarProdutoRequest struct {
	Codigo    string `json:"codigo" example:"SKU-001"`
	Descricao string `json:"descricao" example:"Teclado mecanico"`
	Saldo     int    `json:"saldo" example:"10"`
}

type produtoResponse struct {
	ID              uuid.UUID `json:"id"`
	Codigo          string    `json:"codigo"`
	Descricao       string    `json:"descricao"`
	Saldo           int       `json:"saldo"`
	DataCadastro    time.Time `json:"dataCadastro"`
	DataAtualizacao time.Time `json:"dataAtualizacao"`
}

type errorResponse struct {
	Codigo   string `json:"codigo" example:"PRODUTO_NAO_ENCONTRADO"`
	Mensagem string `json:"mensagem" example:"produto nao encontrado"`
}

type ProdutoHandler struct {
	service *application.Service
}

func NewProdutoHandler(service *application.Service) *ProdutoHandler {
	return &ProdutoHandler{service: service}
}

func (handler *ProdutoHandler) RegisterRoutes(router *gin.Engine) {
	group := router.Group("/produtos")
	group.POST("", handler.criar)
	group.GET("", handler.listar)
	group.GET("/:id", handler.buscarPorID)
	group.GET("/codigo/:codigo", handler.buscarPorCodigo)
}

// criar godoc
// @Summary Cadastra um produto
// @Tags Produtos
// @Accept json
// @Produce json
// @Param produto body criarProdutoRequest true "Dados do produto"
// @Success 201 {object} produtoResponse
// @Failure 400 {object} errorResponse
// @Failure 409 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /produtos [post]
func (handler *ProdutoHandler) criar(c *gin.Context) {
	var request criarProdutoRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		respondError(c, http.StatusBadRequest, "REQUISICAO_INVALIDA", "corpo da requisicao invalido")
		return
	}

	produto, err := handler.service.Criar(c.Request.Context(), application.CriarInput{
		Codigo: request.Codigo, Descricao: request.Descricao, Saldo: request.Saldo,
	})
	if err != nil {
		respondDomainError(c, err)
		return
	}

	c.JSON(http.StatusCreated, toProdutoResponse(produto))
}

// listar godoc
// @Summary Lista produtos
// @Tags Produtos
// @Produce json
// @Success 200 {array} produtoResponse
// @Failure 500 {object} errorResponse
// @Router /produtos [get]
func (handler *ProdutoHandler) listar(c *gin.Context) {
	produtos, err := handler.service.Listar(c.Request.Context())
	if err != nil {
		respondDomainError(c, err)
		return
	}

	response := make([]produtoResponse, 0, len(produtos))
	for index := range produtos {
		response = append(response, toProdutoResponse(&produtos[index]))
	}
	c.JSON(http.StatusOK, response)
}

// buscarPorID godoc
// @Summary Consulta um produto por ID
// @Tags Produtos
// @Produce json
// @Param id path string true "UUID do produto"
// @Success 200 {object} produtoResponse
// @Failure 400 {object} errorResponse
// @Failure 404 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /produtos/{id} [get]
func (handler *ProdutoHandler) buscarPorID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, "ID_INVALIDO", "id do produto deve ser um UUID valido")
		return
	}

	produto, err := handler.service.BuscarPorID(c.Request.Context(), id)
	if err != nil {
		respondDomainError(c, err)
		return
	}
	c.JSON(http.StatusOK, toProdutoResponse(produto))
}

// buscarPorCodigo godoc
// @Summary Consulta um produto por codigo
// @Tags Produtos
// @Produce json
// @Param codigo path string true "Codigo do produto"
// @Success 200 {object} produtoResponse
// @Failure 404 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /produtos/codigo/{codigo} [get]
func (handler *ProdutoHandler) buscarPorCodigo(c *gin.Context) {
	produto, err := handler.service.BuscarPorCodigo(c.Request.Context(), c.Param("codigo"))
	if err != nil {
		respondDomainError(c, err)
		return
	}
	c.JSON(http.StatusOK, toProdutoResponse(produto))
}

func respondDomainError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrCodigoObrigatorio):
		respondError(c, http.StatusBadRequest, "CODIGO_OBRIGATORIO", err.Error())
	case errors.Is(err, domain.ErrDescricaoObrigatoria):
		respondError(c, http.StatusBadRequest, "DESCRICAO_OBRIGATORIA", err.Error())
	case errors.Is(err, domain.ErrSaldoInvalido):
		respondError(c, http.StatusBadRequest, "SALDO_INVALIDO", err.Error())
	case errors.Is(err, domain.ErrCodigoJaExistente):
		respondError(c, http.StatusConflict, "CODIGO_PRODUTO_JA_EXISTENTE", err.Error())
	case errors.Is(err, domain.ErrProdutoNaoEncontrado):
		respondError(c, http.StatusNotFound, "PRODUTO_NAO_ENCONTRADO", err.Error())
	default:
		respondError(c, http.StatusInternalServerError, "ERRO_INTERNO", "erro interno do servidor")
	}
}

func respondError(c *gin.Context, status int, code, message string) {
	c.JSON(status, errorResponse{Codigo: code, Mensagem: message})
}

func toProdutoResponse(produto *domain.Produto) produtoResponse {
	return produtoResponse{
		ID: produto.ID, Codigo: produto.Codigo, Descricao: produto.Descricao,
		Saldo: produto.Saldo, DataCadastro: produto.DataCadastro,
		DataAtualizacao: produto.DataAtualizacao,
	}
}
