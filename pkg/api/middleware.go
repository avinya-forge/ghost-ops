package api

import (
	"net/http"
	"regexp"
)

var serviceIDRegex = regexp.MustCompile(`^[a-zA-Z0-9-]+$`)

// validateServiceIDMiddleware wraps a handler to validate the "id" path value.
func validateServiceIDMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id != "" {
			if !serviceIDRegex.MatchString(id) {
				http.Error(w, "Invalid service ID format", http.StatusBadRequest)
				return
			}
		}
		next(w, r)
	}
}
