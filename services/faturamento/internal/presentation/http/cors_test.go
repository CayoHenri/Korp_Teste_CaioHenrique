package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestCORSMiddlewareAllowsConfiguredOrigin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(corsMiddleware([]string{"http://localhost:4200"}))

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodOptions, "/notas-fiscais", nil)
	request.Header.Set("Origin", "http://localhost:4200")
	request.Header.Set("Access-Control-Request-Method", http.MethodGet)
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("esperava status %d, recebeu %d", http.StatusNoContent, recorder.Code)
	}
	if recorder.Header().Get("Access-Control-Allow-Origin") != "http://localhost:4200" {
		t.Fatalf("origem CORS nao foi liberada")
	}
}

func TestCORSMiddlewareRejectsUnknownOriginPreflight(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(corsMiddleware([]string{"http://localhost:4200"}))

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodOptions, "/notas-fiscais", nil)
	request.Header.Set("Origin", "https://origem-nao-permitida.example")
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("esperava status %d, recebeu %d", http.StatusForbidden, recorder.Code)
	}
}
