package store

import (
	"sync"
	"testing"

	"featureflags/internal/model"
)

func TestCreate(t *testing.T) {
	s := New()
	f := model.Flag{Key: "feature-x", Enabled: true, RolloutPercent: 100}
	if err := s.Create(f); err != nil {
		t.Fatalf("Create returned unexpected error: %v", err)
	}

	got, ok := s.Get("feature-x")
	if !ok {
		t.Fatal("expected flag to exist after Create")
	}
	if got != f {
		t.Fatalf("Get = %+v, want %+v", got, f)
	}
}

func TestCreateDuplicateReturnsErrConflict(t *testing.T) {
	s := New()
	f := model.Flag{Key: "dup", Enabled: true}
	if err := s.Create(f); err != nil {
		t.Fatalf("Create returned unexpected error: %v", err)
	}
	if err := s.Create(f); err != ErrConflict {
		t.Fatalf("Create duplicate = %v, want ErrConflict", err)
	}
}

func TestList(t *testing.T) {
	s := New()
	a := model.Flag{Key: "a"}
	b := model.Flag{Key: "b"}
	if err := s.Create(a); err != nil {
		t.Fatal(err)
	}
	if err := s.Create(b); err != nil {
		t.Fatal(err)
	}

	got := s.List()
	if len(got) != 2 {
		t.Fatalf("List len = %d, want 2", len(got))
	}

	seen := map[string]bool{}
	for _, f := range got {
		seen[f.Key] = true
	}
	if !seen["a"] || !seen["b"] {
		t.Fatalf("List missing keys: %+v", got)
	}
}

func TestGet(t *testing.T) {
	s := New()
	f := model.Flag{Key: "known", Enabled: true}
	if err := s.Create(f); err != nil {
		t.Fatal(err)
	}

	got, ok := s.Get("known")
	if !ok || got != f {
		t.Fatalf("Get = %+v, %v; want %+v, true", got, ok, f)
	}

	if _, ok := s.Get("unknown"); ok {
		t.Fatal("Get on unknown key reported present")
	}
}

func TestUpdate(t *testing.T) {
	s := New()
	f := model.Flag{Key: "upd", Enabled: false}
	if err := s.Create(f); err != nil {
		t.Fatal(err)
	}

	updated := model.Flag{Key: "upd", Enabled: true, Description: "new"}
	got, ok := s.Update("upd", updated)
	if !ok {
		t.Fatal("Update reported missing for existing key")
	}
	if !got.Enabled || got.Description != "new" {
		t.Fatalf("Update = %+v, want enabled with description", got)
	}

	stored, _ := s.Get("upd")
	if stored != got {
		t.Fatalf("stored = %+v, want %+v", stored, got)
	}
}

func TestUpdateUnknown(t *testing.T) {
	s := New()
	if _, ok := s.Update("missing", model.Flag{Key: "missing"}); ok {
		t.Fatal("Update on unknown key reported present")
	}
}

func TestDelete(t *testing.T) {
	s := New()
	if err := s.Create(model.Flag{Key: "del"}); err != nil {
		t.Fatal(err)
	}

	if !s.Delete("del") {
		t.Fatal("Delete reported missing for existing key")
	}
	if _, ok := s.Get("del"); ok {
		t.Fatal("flag still exists after Delete")
	}

	if s.Delete("del") {
		t.Fatal("Delete on already-deleted key reported present")
	}
}

func TestDeleteUnknown(t *testing.T) {
	s := New()
	if s.Delete("missing") {
		t.Fatal("Delete on unknown key reported present")
	}
}

func TestConcurrentAccess(t *testing.T) {
	s := New()

	const writers = 8
	const readers = 16
	const iterations = 200

	var wg sync.WaitGroup
	wg.Add(writers + readers)

	for i := 0; i < writers; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				key := "flag-" + string(rune('a'+j%26))
				if err := s.Create(model.Flag{Key: key, Enabled: j%2 == 0}); err == ErrConflict {
					s.Update(key, model.Flag{Key: key, Enabled: j%2 != 0})
				}
			}
		}(i)
	}

	for i := 0; i < readers; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				_ = s.List()
				key := "flag-" + string(rune('a'+j%26))
				_, _ = s.Get(key)
			}
		}(i)
	}

	wg.Wait()

	if len(s.List()) == 0 {
		t.Fatal("expected flags to exist after concurrent writes")
	}
}
