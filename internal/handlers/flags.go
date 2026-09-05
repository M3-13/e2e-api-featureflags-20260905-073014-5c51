package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"featureflags/internal/model"
	"featureflags/internal/store"
)

const maxBodyBytes = 1 << 20

type createFlagRequest struct {
	Key            string `json:"key"`
	Enabled        *bool  `json:"enabled"`
	Description    string `json:"description"`
	RolloutPercent *int   `json:"rollout_percent"`
}

type updateFlagRequest struct {
	Enabled        bool   `json:"enabled"`
	Description    string `json:"description"`
	RolloutPercent int    `json:"rollout_percent"`
}

func Create(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
		var req createFlagRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			var maxErr *http.MaxBytesError
			if errors.As(err, &maxErr) {
				writeError(w, http.StatusRequestEntityTooLarge, "request body too large")
				return
			}
			writeError(w, http.StatusBadRequest, "invalid JSON body")
			return
		}
		if req.Key == "" {
			writeError(w, http.StatusBadRequest, "key is required")
			return
		}
		if req.Enabled == nil {
			writeError(w, http.StatusBadRequest, "enabled is required")
			return
		}
		rollout := 100
		if req.RolloutPercent != nil {
			rollout = *req.RolloutPercent
		}
		if rollout < 0 || rollout > 100 {
			writeError(w, http.StatusBadRequest, "rollout_percent must be between 0 and 100")
			return
		}
		f := model.Flag{
			Key:            req.Key,
			Enabled:        *req.Enabled,
			Description:    req.Description,
			RolloutPercent: rollout,
		}
		if err := s.Create(f); err != nil {
			if errors.Is(err, store.ErrConflict) {
				writeError(w, http.StatusConflict, "flag already exists")
				return
			}
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}
		writeJSON(w, http.StatusCreated, f)
	}
}

func List(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, s.List())
	}
}

func Get(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := r.PathValue("key")
		f, ok := s.Get(key)
		if !ok {
			writeError(w, http.StatusNotFound, "flag not found")
			return
		}
		writeJSON(w, http.StatusOK, f)
	}
}

func Update(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := r.PathValue("key")
		r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
		var req updateFlagRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			var maxErr *http.MaxBytesError
			if errors.As(err, &maxErr) {
				writeError(w, http.StatusRequestEntityTooLarge, "request body too large")
				return
			}
			writeError(w, http.StatusBadRequest, "invalid JSON body")
			return
		}
		if req.RolloutPercent < 0 || req.RolloutPercent > 100 {
			writeError(w, http.StatusBadRequest, "rollout_percent must be between 0 and 100")
			return
		}
		existing, ok := s.Get(key)
		if !ok {
			writeError(w, http.StatusNotFound, "flag not found")
			return
		}
		updated := model.Flag{
			Key:            existing.Key,
			Enabled:        req.Enabled,
			Description:    req.Description,
			RolloutPercent: req.RolloutPercent,
		}
		f, ok := s.Update(key, updated)
		if !ok {
			writeError(w, http.StatusNotFound, "flag not found")
			return
		}
		writeJSON(w, http.StatusOK, f)
	}
}

func Delete(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := r.PathValue("key")
		if !s.Delete(key) {
			writeError(w, http.StatusNotFound, "flag not found")
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
