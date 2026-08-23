package http

import (
	"net/http"

	application "github.com/caiog/korp-notas-fiscais/services/faturamento/internal/application/nota_fiscal"
	domain "github.com/caiog/korp-notas-fiscais/services/faturamento/internal/domain/nota_fiscal"
	"github.com/caiog/korp-notas-fiscais/services/faturamento/internal/presentation/http/domainerror"
	"github.com/caiog/korp-notas-fiscais/services/faturamento/internal/presentation/http/dto"
	"github.com/caiog/korp-notas-fiscais/services/faturamento/internal/presentation/http/response"
	sharedquery "github.com/caiog/korp-notas-fiscais/services/faturamento/internal/shared/query"
	sharedtext "github.com/caiog/korp-notas-fiscais/services/faturamento/internal/shared/text"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type NotaFiscalHandler struct {
	criar             *application.CriarNotaFiscalUseCase
	atualizar         *application.AtualizarNotaFiscalUseCase
	buscar            *application.BuscarNotaFiscalUseCase
	listar            *application.ListarNotasFiscaisUseCase
	iniciarFechamento *application.IniciarFechamentoUseCase
}

func NewNotaFiscalHandler(
	criar *application.CriarNotaFiscalUseCase,
	atualizar *application.AtualizarNotaFiscalUseCase,
	buscar *application.BuscarNotaFiscalUseCase,
	listar *application.ListarNotasFiscaisUseCase,
	iniciar *application.IniciarFechamentoUseCase,
) *NotaFiscalHandler {
	return &NotaFiscalHandler{
		criar:             criar,
		atualizar:         atualizar,
		buscar:            buscar,
		listar:            listar,
		iniciarFechamento: iniciar,
	}
}

func (handler *NotaFiscalHandler) RegisterRoutes(router *gin.Engine) {
	notas := router.Group("/notas-fiscais")
	notas.POST("", handler.criarNota)
	notas.PUT("/:id", handler.atualizarNota)
	notas.GET("", handler.listarNotas)
	notas.GET("/:id", handler.buscarNota)
	notas.POST("/:id/fechamento", handler.iniciarFechamentoNota)
}

// atualizarNota godoc
// @Summary Atualiza uma nota fiscal aberta
// @Description Substitui cliente, endereco e itens. A operacao e permitida somente no status ABERTA.
// @Tags Notas Fiscais
// @Accept json
// @Produce json
// @Param id path string true "UUID da nota"
// @Param request body dto.AtualizarNotaFiscalRequest true "Novos dados da nota"
// @Success 200 {object} response.SuccessResponse{data=dto.NotaFiscalResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Router /notas-fiscais/{id} [put]
func (handler *NotaFiscalHandler) atualizarNota(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var request dto.AtualizarNotaFiscalRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.Error(c, http.StatusBadRequest, "REQUISICAO_INVALIDA", "dados da requisicao invalidos")
		return
	}
	input := application.AtualizarNotaFiscalInput{
		ID:              id,
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
	nota, err := handler.atualizar.Execute(c.Request.Context(), input)
	if err != nil {
		domainerror.Respond(c, err)
		return
	}
	response.OK(c, dto.NewNotaFiscalResponse(nota))
}

// criarNota godoc
// @Summary Cria uma nota fiscal
// @Tags Notas Fiscais
// @Accept json
// @Produce json
// @Param request body dto.CriarNotaFiscalRequest true "Dados da nota"
// @Success 201 {object} response.SuccessResponse{data=dto.NotaFiscalResponse}
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
// @Param pagina query int false "Pagina" default(1)
// @Param tamanhoPagina query int false "Itens por pagina (maximo 100)" default(20)
// @Param numero query int false "Numero exato da nota"
// @Param status query string false "Status da nota" Enums(ABERTA, PROCESSANDO, FECHADA)
// @Param nomeCliente query string false "Trecho do nome do cliente"
// @Success 200 {object} response.SuccessResponse{data=dto.NotasFiscaisPaginadasResponse}
// @Router /notas-fiscais [get]
func (handler *NotaFiscalHandler) listarNotas(c *gin.Context) {
	var request dto.ListarNotasFiscaisQuery
	if err := c.ShouldBindQuery(&request); err != nil {
		response.Error(c, http.StatusBadRequest, "FILTROS_INVALIDOS", "filtros da requisicao invalidos")
		return
	}
	var status *domain.Status
	if request.Status != "" {
		value := domain.Status(sharedtext.NormalizeUpper(request.Status))
		if !value.Valido() {
			response.Error(c, http.StatusBadRequest, "STATUS_INVALIDO", "status da nota fiscal e invalido")
			return
		}
		status = &value
	}
	pagina, err := handler.listar.Execute(
		c.Request.Context(),
		sharedquery.Criteria[domain.ListFilters]{
			Filters: domain.ListFilters{
				Numero:      request.Numero,
				Status:      status,
				NomeCliente: request.NomeCliente,
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
	result := dto.NotasFiscaisPaginadasResponse{
		Itens:         make([]dto.NotaFiscalResponse, 0, len(pagina.Items)),
		Total:         pagina.Total,
		Pagina:        pagina.Page,
		TamanhoPagina: pagina.PageSize,
		TotalPaginas:  pagina.TotalPages,
	}
	for index := range pagina.Items {
		result.Itens = append(result.Itens, dto.NewNotaFiscalResponse(&pagina.Items[index]))
	}
	response.OK(c, result)
}

// buscarNota godoc
// @Summary Consulta uma nota fiscal
// @Tags Notas Fiscais
// @Produce json
// @Param id path string true "UUID da nota"
// @Success 200 {object} response.SuccessResponse{data=dto.NotaFiscalResponse}
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
// @Success 200 {object} response.SuccessResponse{data=dto.NotaFiscalResponse}
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
