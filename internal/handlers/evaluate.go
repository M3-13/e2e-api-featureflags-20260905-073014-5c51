package handlers

import (
	"net/http"

	"featureflags/internal/hash"
	"featureflags/internal/model"
	"featureflags/internal/store"
)

// evaluateFlag computes the deterministic rollout decision for a single flag
// and user. A disabled flag always yields false; otherwise the user is mapped
// to a bucket in [0, 99] and compared against the rollout percentage.
func evaluateFlag(flag model.Flag, user string) bool {
	if !flag.Enabled {
		return false
	}
	return hash.Bucket(flag.Key, user) < flag.RolloutPercent
}

// evaluateResponse builds the JSON body returned by the evaluate endpoint.
func evaluateResponse(flag model.Flag, user string) map[string]any {
	return map[string]any{
		"key":     flag.Key,
		"enabled": flag.Enabled,
		"result":  evaluateFlag(flag, user),
	}
}

func Evaluate(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := r.PathValue("key")
		user := r.URL.Query().Get("user")

		if user == "" {
			writeError(w, http.StatusBadRequest, "user is required")
			return
		}

		flag, ok := s.Get(key)
		if !ok {
			writeError(w, http.StatusNotFound, "flag not found")
			return
		}

		writeJSON(w, http.StatusOK, evaluateResponse(flag, user))
	}
}
