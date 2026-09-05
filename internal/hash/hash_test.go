package hash

import "testing"

func TestBucketDeterministic(t *testing.T) {
	key := "feature-x"
	user := "alice"
	first := Bucket(key, user)
	for i := 0; i < 100; i++ {
		if got := Bucket(key, user); got != first {
			t.Fatalf("Bucket(%q, %q) not deterministic: %d then %d", key, user, first, got)
		}
	}
}

func TestBucketRange(t *testing.T) {
	keys := []string{"a", "feature-x", "a-longer-flag-key-name", ""}
	users := []string{"alice", "bob", "", "user@example.com", "日本語"}
	for _, k := range keys {
		for _, u := range users {
			if got := Bucket(k, u); got < 0 || got > 99 {
				t.Errorf("Bucket(%q, %q) = %d, out of range [0, 99]", k, u, got)
			}
		}
	}
}
