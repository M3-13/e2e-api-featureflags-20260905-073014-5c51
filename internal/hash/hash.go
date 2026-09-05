package hash

const (
	fnvOffset64 = 14695981039346656037
	fnvPrime64  = 1099511628211
)

func fnv1a(data string) uint64 {
	h := uint64(fnvOffset64)
	for i := 0; i < len(data); i++ {
		h ^= uint64(data[i])
		h *= fnvPrime64
	}
	return h
}

// Bucket returns a deterministic value in [0, 99] for the given key and user.
// The user value is only ever used transiently here for hashing and is never
// persisted.
func Bucket(key, user string) int {
	h := fnv1a(key)
	h ^= fnv1a(user)
	return int(h % 100)
}
