package http

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

type databaseStub struct {
	err error
}

func (stub databaseStub) PingContext(context.Context) error {
	return stub.err
}

func TestHealthReturnsOKWhenDatabaseIsAvailable(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/health", nil)

	NewRouter(databaseStub{}, nil).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("esperava status %d, recebeu %d", http.StatusOK, recorder.Code)
	}
}

func TestHealthReturnsServiceUnavailableWhenDatabaseFails(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/health", nil)

	NewRouter(databaseStub{err: errors.New("database unavailable")}, nil).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("esperava status %d, recebeu %d", http.StatusServiceUnavailable, recorder.Code)
	}
}
