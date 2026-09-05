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
	return nil
}

func (s *Store) List() []model.Flag {
	return nil
}

func (s *Store) Get(key string) (model.Flag, bool) {
	return model.Flag{}, false
}

func (s *Store) Update(key string, f model.Flag) (model.Flag, bool) {
	return model.Flag{}, false
}

func (s *Store) Delete(key string) bool {
	return false
}
