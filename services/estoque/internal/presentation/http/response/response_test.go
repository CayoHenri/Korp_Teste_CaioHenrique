package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestCreatedWrapsDataInSuccessEnvelope(t *testing.T) {
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)

	Created(context, map[string]string{"id": "123"})

	var body map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("resposta JSON invalida: %v", err)
	}
	if recorder.Code != http.StatusCreated || body["success"] != true || body["data"] == nil {
		t.Fatalf("envelope de sucesso inesperado: status=%d body=%v", recorder.Code, body)
	}
}

func TestErrorHidesDataAndReturnsErrorEnvelope(t *testing.T) {
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)

	Error(context, http.StatusNotFound, "NAO_ENCONTRADO", "recurso nao encontrado")

	var body map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("resposta JSON invalida: %v", err)
	}
	if recorder.Code != http.StatusNotFound || body["success"] != false || body["error"] == nil {
		t.Fatalf("envelope de erro inesperado: status=%d body=%v", recorder.Code, body)
	}
	if _, exists := body["data"]; exists {
		t.Fatalf("resposta de erro nao deve conter data: %v", body)
	}
}
