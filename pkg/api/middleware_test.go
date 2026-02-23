package api

import (
	"net/http"
	"net/http/httptest"
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
