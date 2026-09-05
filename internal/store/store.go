package store

import (
	"errors"
	"sync"

	"featureflags/internal/model"
)

var ErrConflict = errors.New("flag already exists")

type Store struct {
	mu    sync.RWMutex
	flags map[string]model.Flag
}

func New() *Store {
	return &Store{
		flags: make(map[string]model.Flag),
	}
}

func (s *Store) Create(f model.Flag) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.flags[f.Key]; exists {
		return ErrConflict
	}
	s.flags[f.Key] = f
	return nil
}

func (s *Store) List() []model.Flag {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]model.Flag, 0, len(s.flags))
	for _, f := range s.flags {
		out = append(out, f)
	}
	return out
}

func (s *Store) Get(key string) (model.Flag, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	f, ok := s.flags[key]
	return f, ok
}

func (s *Store) Update(key string, f model.Flag) (model.Flag, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.flags[key]; !exists {
		return model.Flag{}, false
	}
	f.Key = key
	s.flags[key] = f
	return f, true
}

func (s *Store) Delete(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.flags[key]; !exists {
		return false
	}
	delete(s.flags, key)
	return true
}
