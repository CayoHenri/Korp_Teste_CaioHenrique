package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/joho/godotenv"
)

type testConfig struct {
	estoqueURL     string
	faturamentoURL string
	timeout        time.Duration
}

type successResponse[T any] struct {
	Success bool `json:"success"`
	Data    T    `json:"data"`
}

type produtoResponse struct {
	ID     string `json:"id"`
	Codigo string `json:"codigo"`
	Saldo  int    `json:"saldo"`
}

type notaFiscalResponse struct {
	ID             string `json:"id"`
	Status         string `json:"status"`
	MotivoRejeicao string `json:"motivoRejeicao"`
}

type movimentacaoResponse struct {
	Tipo       string  `json:"tipo"`
	Quantidade int     `json:"quantidade"`
	Referencia *string `json:"referencia"`
}

func TestFluxoFechamentoComSaldoSuficiente(t *testing.T) {
	cfg := loadConfig(t)
	client := &http.Client{Timeout: 5 * time.Second}
	codigo := uniqueCode("SUCESSO")
	produto := createProduct(t, client, cfg, codigo, 10)
	nota := createInvoice(t, client, cfg, codigo, 3)
	startClosing(t, client, cfg, nota.ID)

	final := waitForStatus(t, client, cfg, nota.ID, "FECHADA")
	if final.MotivoRejeicao != "" {
		t.Fatalf("nota fechada nao deveria possuir motivo: %s", final.MotivoRejeicao)
	}
	produto = getProduct(t, client, cfg, codigo)
	if produto.Saldo != 7 {
		t.Fatalf("esperava saldo 7, recebeu %d", produto.Saldo)
	}
	movements := getMovements(t, client, cfg, produto.ID)
	found := false
	for _, movement := range movements {
		if movement.Tipo == "SAIDA" && movement.Quantidade == 3 &&
			movement.Referencia != nil && *movement.Referencia == nota.ID {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("movimentacao de saida vinculada a nota nao foi encontrada")
	}
}

func TestFluxoFechamentoRejeitadoPorSaldoInsuficiente(t *testing.T) {
	cfg := loadConfig(t)
	client := &http.Client{Timeout: 5 * time.Second}
	codigo := uniqueCode("REJEICAO")
	createProduct(t, client, cfg, codigo, 1)
	nota := createInvoice(t, client, cfg, codigo, 5)
	startClosing(t, client, cfg, nota.ID)

	final := waitForRejection(t, client, cfg, nota.ID)
	if final.MotivoRejeicao != "ESTOQUE_INSUFICIENTE" {
		t.Fatalf("motivo inesperado: %s", final.MotivoRejeicao)
	}
	produto := getProduct(t, client, cfg, codigo)
	if produto.Saldo != 1 {
		t.Fatalf("saldo da baixa rejeitada foi alterado: %d", produto.Saldo)
	}
}

func loadConfig(t *testing.T) testConfig {
	t.Helper()
	loadRootEnvironment(t)
	estoqueURL := requiredEnvironment(t, "E2E_ESTOQUE_URL")
	faturamentoURL := requiredEnvironment(t, "E2E_FATURAMENTO_URL")
	timeoutValue := requiredEnvironment(t, "E2E_TIMEOUT")
	timeout, err := time.ParseDuration(timeoutValue)
	if err != nil || timeout <= 0 {
		t.Fatalf("E2E_TIMEOUT deve ser uma duracao positiva: %q", timeoutValue)
	}
	return testConfig{
		estoqueURL:     strings.TrimRight(estoqueURL, "/"),
		faturamentoURL: strings.TrimRight(faturamentoURL, "/"),
		timeout:        timeout,
	}
}

func loadRootEnvironment(t *testing.T) {
	t.Helper()
	directory, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		path := filepath.Join(directory, ".env")
		if _, err := os.Stat(path); err == nil {
			if err := godotenv.Load(path); err != nil {
				t.Fatal(err)
			}
			return
		}
		parent := filepath.Dir(directory)
		if parent == directory {
			return
		}
		directory = parent
	}
}

func requiredEnvironment(t *testing.T, key string) string {
	t.Helper()
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		t.Fatalf("variavel de ambiente obrigatoria ausente: %s", key)
	}
	return value
}

func uniqueCode(scenario string) string {
	return fmt.Sprintf("E2E-%s-%d", scenario, time.Now().UnixNano())
}

func createProduct(
	t *testing.T,
	client *http.Client,
	cfg testConfig,
	code string,
	balance int,
) produtoResponse {
	t.Helper()
	return request[produtoResponse](t, client, http.MethodPost, cfg.estoqueURL+"/produtos", map[string]any{
		"codigo": code, "descricao": "Produto E2E", "saldo": balance, "valor": 25.50,
	}, http.StatusCreated)
}

func createInvoice(
	t *testing.T,
	client *http.Client,
	cfg testConfig,
	productCode string,
	quantity int,
) notaFiscalResponse {
	t.Helper()
	return request[notaFiscalResponse](t, client, http.MethodPost, cfg.faturamentoURL+"/notas-fiscais", map[string]any{
		"nomeCliente": "Cliente E2E", "enderecoCliente": "Rua E2E, 100",
		"itens": []map[string]any{{"codigoProduto": productCode, "quantidade": quantity}},
	}, http.StatusCreated)
}

func startClosing(t *testing.T, client *http.Client, cfg testConfig, invoiceID string) {
	t.Helper()
	result := request[notaFiscalResponse](
		t, client, http.MethodPost,
		cfg.faturamentoURL+"/notas-fiscais/"+invoiceID+"/fechamento",
		nil, http.StatusOK,
	)
	if result.Status != "PROCESSANDO" {
		t.Fatalf("esperava PROCESSANDO, recebeu %s", result.Status)
	}
}

func getProduct(t *testing.T, client *http.Client, cfg testConfig, code string) produtoResponse {
	t.Helper()
	return request[produtoResponse](
		t, client, http.MethodGet, cfg.estoqueURL+"/produtos/codigo/"+code, nil, http.StatusOK,
	)
}

func getInvoice(t *testing.T, client *http.Client, cfg testConfig, id string) notaFiscalResponse {
	t.Helper()
	return request[notaFiscalResponse](
		t, client, http.MethodGet, cfg.faturamentoURL+"/notas-fiscais/"+id, nil, http.StatusOK,
	)
}

func getMovements(
	t *testing.T,
	client *http.Client,
	cfg testConfig,
	productID string,
) []movimentacaoResponse {
	t.Helper()
	return request[[]movimentacaoResponse](
		t, client, http.MethodGet,
		cfg.estoqueURL+"/produtos/"+productID+"/movimentacoes",
		nil, http.StatusOK,
	)
}

func waitForStatus(
	t *testing.T,
	client *http.Client,
	cfg testConfig,
	invoiceID, expected string,
) notaFiscalResponse {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), cfg.timeout)
	defer cancel()
	for {
		invoice := getInvoice(t, client, cfg, invoiceID)
		if invoice.Status == expected {
			return invoice
		}
		select {
		case <-ctx.Done():
			t.Fatalf("nota %s permaneceu no status %s", invoiceID, invoice.Status)
		case <-time.After(250 * time.Millisecond):
		}
	}
}

func waitForRejection(
	t *testing.T,
	client *http.Client,
	cfg testConfig,
	invoiceID string,
) notaFiscalResponse {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), cfg.timeout)
	defer cancel()
	for {
		invoice := getInvoice(t, client, cfg, invoiceID)
		if invoice.Status == "ABERTA" && invoice.MotivoRejeicao != "" {
			return invoice
		}
		select {
		case <-ctx.Done():
			t.Fatalf("nota %s nao recebeu rejeicao; status=%s", invoiceID, invoice.Status)
		case <-time.After(250 * time.Millisecond):
		}
	}
}

func request[T any](
	t *testing.T,
	client *http.Client,
	method, url string,
	body any,
	expectedStatus int,
) T {
	t.Helper()
	var reader io.Reader
	if body != nil {
		content, err := json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
		reader = bytes.NewReader(content)
	}
	req, err := http.NewRequest(method, url, reader)
	if err != nil {
		t.Fatal(err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	response, err := client.Do(req)
	if err != nil {
		t.Fatalf("%s %s: %v", method, url, err)
	}
	defer response.Body.Close()
	content, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != expectedStatus {
		t.Fatalf("%s %s retornou %d: %s", method, url, response.StatusCode, content)
	}
	var envelope successResponse[T]
	if err := json.Unmarshal(content, &envelope); err != nil {
		t.Fatalf("resposta invalida de %s %s: %v; body=%s", method, url, err, content)
	}
	if !envelope.Success {
		t.Fatalf("%s %s retornou success=false: %s", method, url, content)
	}
	return envelope.Data
}
