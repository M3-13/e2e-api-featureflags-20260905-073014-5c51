package handlers

import (
	"net/http"

	"featureflags/internal/store"
)

func Evaluate(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeError(w, http.StatusNotImplemented, "not implemented")
	}
}
