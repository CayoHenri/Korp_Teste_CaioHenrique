package http

import (
	application "github.com/caiog/korp-notas-fiscais/services/faturamento/internal/application/nota_fiscal"
	"github.com/caiog/korp-notas-fiscais/services/faturamento/internal/presentation/http/domainerror"
	"github.com/caiog/korp-notas-fiscais/services/faturamento/internal/presentation/http/dto"
	"github.com/caiog/korp-notas-fiscais/services/faturamento/internal/presentation/http/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
)

type NotaFiscalHandler struct {
	criar             *application.CriarNotaFiscalUseCase
	buscar            *application.BuscarNotaFiscalUseCase
	listar            *application.ListarNotasFiscaisUseCase
	iniciarFechamento *application.IniciarFechamentoUseCase
}

func NewNotaFiscalHandler(criar *application.CriarNotaFiscalUseCase, buscar *application.BuscarNotaFiscalUseCase, listar *application.ListarNotasFiscaisUseCase, iniciar *application.IniciarFechamentoUseCase) *NotaFiscalHandler {
	return &NotaFiscalHandler{criar: criar, buscar: buscar, listar: listar, iniciarFechamento: iniciar}
}
func (handler *NotaFiscalHandler) RegisterRoutes(router *gin.Engine) {
	notas := router.Group("/notas-fiscais")
	notas.POST("", handler.criarNota)
	notas.GET("", handler.listarNotas)
	notas.GET("/:id", handler.buscarNota)
	notas.POST("/:id/fechamento", handler.iniciarFechamentoNota)
}

// criarNota godoc
// @Summary Cria uma nota fiscal
// @Tags Notas Fiscais
// @Accept json
// @Produce json
// @Param request body dto.CriarNotaFiscalRequest true "Dados da nota"
// @Success 201 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Router /notas-fiscais [post]
func (handler *NotaFiscalHandler) criarNota(c *gin.Context) {
	var request dto.CriarNotaFiscalRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.Error(c, http.StatusBadRequest, "REQUISICAO_INVALIDA", "dados da requisicao invalidos")
		return
	}
	input := application.CriarNotaFiscalInput{
		NomeCliente:     request.NomeCliente,
		EnderecoCliente: request.EnderecoCliente,
		Itens:           make([]application.CriarNotaFiscalItemInput, 0, len(request.Itens)),
	}
	for _, item := range request.Itens {
		input.Itens = append(input.Itens, application.CriarNotaFiscalItemInput{
			CodigoProduto: item.CodigoProduto,
			Quantidade:    item.Quantidade,
		})
	}
	nota, err := handler.criar.Execute(c.Request.Context(), input)
	if err != nil {
		domainerror.Respond(c, err)
		return
	}
	response.Created(c, dto.NewNotaFiscalResponse(nota))
}

// listarNotas godoc
// @Summary Lista notas fiscais
// @Tags Notas Fiscais
// @Produce json
// @Success 200 {object} response.SuccessResponse
// @Router /notas-fiscais [get]
func (handler *NotaFiscalHandler) listarNotas(c *gin.Context) {
	notas, err := handler.listar.Execute(c.Request.Context())
	if err != nil {
		domainerror.Respond(c, err)
		return
	}
	result := make([]dto.NotaFiscalResponse, 0, len(notas))
	for index := range notas {
		result = append(result, dto.NewNotaFiscalResponse(&notas[index]))
	}
	response.OK(c, result)
}

// buscarNota godoc
// @Summary Consulta uma nota fiscal
// @Tags Notas Fiscais
// @Produce json
// @Param id path string true "UUID da nota"
// @Success 200 {object} response.SuccessResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /notas-fiscais/{id} [get]
func (handler *NotaFiscalHandler) buscarNota(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	nota, err := handler.buscar.Execute(c.Request.Context(), id)
	if err != nil {
		domainerror.Respond(c, err)
		return
	}
	response.OK(c, dto.NewNotaFiscalResponse(nota))
}

// iniciarFechamentoNota godoc
// @Summary Inicia o fechamento de uma nota fiscal
// @Tags Notas Fiscais
// @Produce json
// @Param id path string true "UUID da nota"
// @Success 200 {object} response.SuccessResponse
// @Failure 409 {object} response.ErrorResponse
// @Router /notas-fiscais/{id}/fechamento [post]
func (handler *NotaFiscalHandler) iniciarFechamentoNota(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	nota, err := handler.iniciarFechamento.Execute(c.Request.Context(), id)
	if err != nil {
		domainerror.Respond(c, err)
		return
	}
	response.OK(c, dto.NewNotaFiscalResponse(nota))
}

func parseID(c *gin.Context) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "ID_INVALIDO", "id deve ser um UUID valido")
		return uuid.Nil, false
	}
	return id, true
}
