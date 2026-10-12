package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSynthesizedMiddleware(t *testing.T) {
	next := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}

	handler := SynthesizedMiddleware(next)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if h := w.Header().Get("X-Ghost-Ops-Synthesized"); h != "true" {
		t.Errorf("Expected X-Ghost-Ops-Synthesized: true, got %q", h)
	}
}
