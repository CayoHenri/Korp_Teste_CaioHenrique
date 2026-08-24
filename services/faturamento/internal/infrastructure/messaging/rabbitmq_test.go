package messaging

import (
	"testing"

	amqp "github.com/rabbitmq/amqp091-go"
)

func TestRetryCountAceitaTiposNumericosDoAMQP(t *testing.T) {
	tests := []struct {
		value any
		want  int
	}{{int32(2), 2}, {int64(3), 3}, {4, 4}, {"5", 0}}
	for _, test := range tests {
		if got := retryCount(amqp.Table{"x-retry-count": test.value}); got != test.want {
			t.Fatalf("retryCount(%v) = %d; esperado %d", test.value, got, test.want)
		}
	}
}

func TestCopiarMensagemNaoAlteraHeadersOriginais(t *testing.T) {
	delivery := amqp.Delivery{Headers: amqp.Table{"original": "valor"}}
	message := copiarMensagem(delivery)
	message.Headers["novo"] = "valor"
	if _, exists := delivery.Headers["novo"]; exists {
		t.Fatal("a copia compartilhou o mapa de headers com a entrega")
	}
}
