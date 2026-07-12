package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAppHandlerServesOnlyEmbeddedIndex(t *testing.T) {
	t.Run("serves app index", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/app/", nil)
		response := httptest.NewRecorder()

		appHandler(response, request)

		if response.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
		}
		if !strings.Contains(response.Body.String(), "Welcome to Chirpy") {
			t.Fatalf("expected embedded index, got %q", response.Body.String())
		}
	})

	for _, path := range []string{"/app/.env", "/app/go.mod", "/app/../.env"} {
		t.Run("rejects "+path, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, path, nil)
			response := httptest.NewRecorder()

			appHandler(response, request)

			if response.Code != http.StatusNotFound {
				t.Fatalf("expected status %d for %s, got %d", http.StatusNotFound, path, response.Code)
			}
		})
	}
}
