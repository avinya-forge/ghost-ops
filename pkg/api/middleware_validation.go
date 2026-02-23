package api

import (
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"
	"strings"
)

// MaxBytesMiddleware enforces a maximum request body size.
func MaxBytesMiddleware(limit int64) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if r.ContentLength > limit {
				http.Error(w, "Request entity too large", http.StatusRequestEntityTooLarge)
				return
			}
			r.Body = http.MaxBytesReader(w, r.Body, limit)
			next(w, r)
		}
	}
}

// ContentTypeMiddleware enforces the Content-Type header for requests with a body.
func ContentTypeMiddleware(allowed ...string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Only check Content-Type for methods that typically have a body
			if r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodPatch {
				ct := r.Header.Get("Content-Type")
				if ct == "" {
					http.Error(w, "Content-Type header is required", http.StatusBadRequest)
					return
				}

				valid := false
				for _, t := range allowed {
					if strings.HasPrefix(ct, t) {
						valid = true
						break
					}
				}

				if !valid {
					http.Error(w, fmt.Sprintf("Unsupported Content-Type. Allowed: %v", allowed), http.StatusUnsupportedMediaType)
					return
				}
			}
			next(w, r)
		}
	}
}

// RecoverMiddleware recovers from panics and logs the error.
func RecoverMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				slog.Error("Panic recovered", "error", err, "stack", string(debug.Stack()))
				http.Error(w, "Internal server error", http.StatusInternalServerError)
			}
		}()
		next(w, r)
	}
}
