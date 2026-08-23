package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	application "github.com/caiog/korp-notas-fiscais/services/faturamento/internal/application/nota_fiscal"
	sharedtext "github.com/caiog/korp-notas-fiscais/services/faturamento/internal/shared/text"
)

type EstoqueClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewEstoqueClient(baseURL string, httpClient *http.Client) *EstoqueClient {
	return &EstoqueClient{
		baseURL:    strings.TrimRight(baseURL, "/"),
		httpClient: httpClient,
	}
}

func (client *EstoqueClient) BuscarPorCodigo(
	ctx context.Context,
	codigo string,
) (*application.ProdutoCatalogo, error) {
	codigoNormalizado := sharedtext.NormalizeUpper(codigo)
	endpoint := client.baseURL + "/produtos/codigo/" + url.PathEscape(codigoNormalizado)
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("criar requisicao ao estoque: %w", err)
	}
	result, err := client.httpClient.Do(request)
	if err != nil {
		return nil, application.ErrEstoqueIndisponivel
	}
	defer result.Body.Close()

	if result.StatusCode == http.StatusNotFound {
		return nil, application.ErrProdutoNaoEncontrado
	}
	if result.StatusCode != http.StatusOK {
		return nil, application.ErrEstoqueIndisponivel
	}
	var envelope struct {
		Data application.ProdutoCatalogo `json:"data"`
	}

	if err := json.NewDecoder(result.Body).Decode(&envelope); err != nil {
		return nil, application.ErrEstoqueIndisponivel
	}
	return &envelope.Data, nil
}
