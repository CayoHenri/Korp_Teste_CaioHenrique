package e2e

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	eventsExchange    = "korp.events"
	requestRoutingKey = "estoque.baixa.solicitada"
	successRoutingKey = "estoque.baixa.realizada"
	requestDLQ        = "estoque.baixa.solicitada.dlq"
	resultDLQ         = "faturamento.baixa.resultado.dlq"
)

func TestFechamentoRejeitaProdutoInativadoDepoisDaCriacao(t *testing.T) {
	cfg := loadConfig(t)
	client := &http.Client{Timeout: 5 * time.Second}
	code := uniqueCode("INATIVO")
	product := createProduct(t, client, cfg, code, 10)
	invoice := createInvoice(t, client, cfg, code, 2)
	request[map[string]any](
		t, client, http.MethodPatch,
		cfg.estoqueURL+"/produtos/"+product.ID+"/inativar",
		nil, http.StatusOK,
	)
	startClosing(t, client, cfg, invoice.ID)
	final := waitForRejection(t, client, cfg, invoice.ID)
	if final.MotivoRejeicao != "PRODUTO_INATIVO" {
		t.Fatalf("motivo inesperado: %s", final.MotivoRejeicao)
	}
	if getProduct(t, client, cfg, code).Saldo != 10 {
		t.Fatal("produto inativo teve o saldo alterado")
	}
}

func TestBaixaDeMultiplosItensFazRollbackCompleto(t *testing.T) {
	cfg := loadConfig(t)
	client := &http.Client{Timeout: 5 * time.Second}
	firstCode := uniqueCode("ATOMICO-A")
	secondCode := uniqueCode("ATOMICO-B")
	createProduct(t, client, cfg, firstCode, 10)
	createProduct(t, client, cfg, secondCode, 1)
	invoice := createInvoiceWithItems(t, client, cfg, []map[string]any{
		{"codigoProduto": firstCode, "quantidade": 3},
		{"codigoProduto": secondCode, "quantidade": 5},
	})
	startClosing(t, client, cfg, invoice.ID)
	final := waitForRejection(t, client, cfg, invoice.ID)
	if final.MotivoRejeicao != "ESTOQUE_INSUFICIENTE" {
		t.Fatalf("motivo inesperado: %s", final.MotivoRejeicao)
	}
	if getProduct(t, client, cfg, firstCode).Saldo != 10 ||
		getProduct(t, client, cfg, secondCode).Saldo != 1 {
		t.Fatal("a baixa parcial nao foi revertida")
	}
}

func TestSolicitacaoDuplicadaNaoBaixaSaldoDuasVezes(t *testing.T) {
	cfg := loadConfig(t)
	client := &http.Client{Timeout: 5 * time.Second}
	product := createProduct(t, client, cfg, uniqueCode("DUPLICADA"), 10)
	eventID := uuid.New()
	payload, err := json.Marshal(map[string]any{
		"eventId": eventID, "notaFiscalId": uuid.New(),
		"itens": []map[string]any{{"produtoId": product.ID, "quantidade": 2}},
	})
	if err != nil {
		t.Fatal(err)
	}
	publish(t, cfg, requestRoutingKey, eventID.String(), payload)
	publish(t, cfg, requestRoutingKey, eventID.String(), payload)
	waitUntil(t, cfg.timeout, func() bool {
		return getProduct(t, client, cfg, product.Codigo).Saldo == 8
	}, "saldo nao foi atualizado pela solicitacao duplicada")
	time.Sleep(500 * time.Millisecond)
	if balance := getProduct(t, client, cfg, product.Codigo).Saldo; balance != 8 {
		t.Fatalf("solicitacao duplicada alterou o saldo novamente: %d", balance)
	}
}

func TestResultadoDuplicadoNaoRepeteTransicao(t *testing.T) {
	cfg := loadConfig(t)
	client := &http.Client{Timeout: 5 * time.Second}
	code := uniqueCode("RESULTADO-DUPLICADO")
	createProduct(t, client, cfg, code, 10)
	invoice := createInvoice(t, client, cfg, code, 1)
	startClosing(t, client, cfg, invoice.ID)
	waitForStatus(t, client, cfg, invoice.ID, "FECHADA")
	database := openDatabase(t, cfg)
	correlationID := queryCorrelationID(t, database, invoice.ID)
	before := countProcessedCorrelation(t, database, correlationID)
	eventID := uuid.New()
	payload, _ := json.Marshal(map[string]any{
		"eventId": eventID, "correlationId": correlationID, "notaFiscalId": invoice.ID,
	})
	publish(t, cfg, successRoutingKey, eventID.String(), payload)
	time.Sleep(750 * time.Millisecond)
	if getInvoice(t, client, cfg, invoice.ID).Status != "FECHADA" {
		t.Fatal("resultado duplicado alterou o estado final da nota")
	}
	if after := countProcessedCorrelation(t, database, correlationID); after != before || after != 1 {
		t.Fatalf("idempotencia inesperada: antes=%d depois=%d", before, after)
	}
}

func TestMensagensInvalidasSaoEncaminhadasParaDLQ(t *testing.T) {
	cfg := loadConfig(t)
	requestBefore := queueMessages(t, cfg, requestDLQ)
	resultBefore := queueMessages(t, cfg, resultDLQ)
	publish(t, cfg, requestRoutingKey, uuid.NewString(), []byte(`{"json":`))
	publish(t, cfg, successRoutingKey, uuid.NewString(), []byte(`{"json":`))
	waitUntil(t, cfg.timeout, func() bool {
		return queueMessages(t, cfg, requestDLQ) >= requestBefore+1 &&
			queueMessages(t, cfg, resultDLQ) >= resultBefore+1
	}, "mensagens invalidas nao chegaram as DLQs")
}

func createInvoiceWithItems(
	t *testing.T,
	client *http.Client,
	cfg testConfig,
	items []map[string]any,
) notaFiscalResponse {
	t.Helper()
	return request[notaFiscalResponse](
		t, client, http.MethodPost, cfg.faturamentoURL+"/notas-fiscais",
		map[string]any{
			"nomeCliente": "Cliente E2E", "enderecoCliente": "Rua E2E, 100", "itens": items,
		},
		http.StatusCreated,
	)
}

func publish(t *testing.T, cfg testConfig, routingKey, messageID string, body []byte) {
	t.Helper()
	connection, err := amqp.Dial(cfg.rabbitMQURL)
	if err != nil {
		t.Fatal(err)
	}
	defer connection.Close()
	channel, err := connection.Channel()
	if err != nil {
		t.Fatal(err)
	}
	defer channel.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := channel.PublishWithContext(
		ctx, eventsExchange, routingKey, false, false,
		amqp.Publishing{
			ContentType: "application/json", DeliveryMode: amqp.Persistent,
			MessageId: messageID, Type: routingKey, Body: body,
		},
	); err != nil {
		t.Fatal(err)
	}
}

func queueMessages(t *testing.T, cfg testConfig, queueName string) int {
	t.Helper()
	connection, err := amqp.Dial(cfg.rabbitMQURL)
	if err != nil {
		t.Fatal(err)
	}
	defer connection.Close()
	channel, err := connection.Channel()
	if err != nil {
		t.Fatal(err)
	}
	defer channel.Close()
	queue, err := channel.QueueInspect(queueName)
	if err != nil {
		t.Fatal(err)
	}
	return queue.Messages
}

func openDatabase(t *testing.T, cfg testConfig) *sql.DB {
	t.Helper()
	database, err := sql.Open("pgx", cfg.databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = database.Close() })
	if err := database.Ping(); err != nil {
		t.Fatal(err)
	}
	return database
}

func queryCorrelationID(t *testing.T, database *sql.DB, invoiceID string) string {
	t.Helper()
	var correlationID string
	if err := database.QueryRow(
		`SELECT id::text FROM faturamento.outbox_events
		WHERE aggregate_id = $1 ORDER BY created_at DESC LIMIT 1`,
		invoiceID,
	).Scan(&correlationID); err != nil {
		t.Fatal(err)
	}
	return correlationID
}

func countProcessedCorrelation(t *testing.T, database *sql.DB, correlationID string) int {
	t.Helper()
	var count int
	if err := database.QueryRow(
		"SELECT COUNT(*) FROM faturamento.mensagens_processadas WHERE correlation_id = $1",
		correlationID,
	).Scan(&count); err != nil {
		t.Fatal(err)
	}
	return count
}

func waitUntil(t *testing.T, timeout time.Duration, condition func() bool, message string) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if condition() {
			return
		}
		time.Sleep(250 * time.Millisecond)
	}
	t.Fatal(message)
}
