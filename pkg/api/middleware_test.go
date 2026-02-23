package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestValidateServiceIDMiddleware(t *testing.T) {
	// Dummy handler that always returns OK
	next := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}

	handler := validateServiceIDMiddleware(next)

	tests := []struct {
		name       string
		pathID     string
		wantStatus int
	}{
		{"Valid ID", "valid-id-123", http.StatusOK},
		{"Valid ID Alphanumeric", "Service1", http.StatusOK},
		{"Invalid ID Space", "invalid id", http.StatusBadRequest},
		{"Invalid ID Special Char", "invalid@id", http.StatusBadRequest},
		{"Invalid ID Path Traversal", "../etc/passwd", http.StatusBadRequest},
		{"Empty ID", "", http.StatusOK}, // Empty ID means middleware sees nothing to validate (e.g. root path)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/services/dummy/logs", nil)
			req.SetPathValue("id", tt.pathID)

			w := httptest.NewRecorder()
			handler(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("validateServiceIDMiddleware() status = %v, want %v", w.Code, tt.wantStatus)
			}
		})
	}
}

func TestMaxBytesMiddleware(t *testing.T) {
	limit := int64(10)
	mw := MaxBytesMiddleware(limit)

	next := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}

	handler := mw(next)

	t.Run("Small Body", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("small"))
		w := httptest.NewRecorder()
		handler(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
	})

	t.Run("Large Body ContentLength", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("too large body"))
		w := httptest.NewRecorder()
		handler(w, req)
		if w.Code != http.StatusRequestEntityTooLarge {
			t.Errorf("Expected status 413, got %d", w.Code)
		}
	})
}

func TestContentTypeMiddleware(t *testing.T) {
	mw := ContentTypeMiddleware("application/json")

	next := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}

	handler := mw(next)

	t.Run("Valid Content-Type", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		handler(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
	})

	t.Run("Invalid Content-Type", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		req.Header.Set("Content-Type", "text/plain")
		w := httptest.NewRecorder()
		handler(w, req)
		if w.Code != http.StatusUnsupportedMediaType {
			t.Errorf("Expected status 415, got %d", w.Code)
		}
	})

	t.Run("Missing Content-Type", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		w := httptest.NewRecorder()
		handler(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})

	t.Run("GET Request Ignored", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()
		handler(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
	})
}

func TestRecoverMiddleware(t *testing.T) {
	mw := RecoverMiddleware

	next := func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	}

	handler := mw(next)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	// Capture slog output to avoid noise? Or just ensure it recovers.
	// We just ensure it returns 500 and doesn't crash the test runner.
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Middleware did not recover from panic")
			}
		}()
		handler(w, req)
	}()

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}
}
