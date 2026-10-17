package api

import (
	"net/http"
)

// SynthesizedMiddleware injects the X-Ghost-Ops-Synthesized header for AI generated services.
// This is required for EU AI Act Art 13 compliance.
func SynthesizedMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Ghost-Ops-Synthesized", "true")
		next(w, r)
	}
}
