package hasher

import (
	"github.com/cespare/xxhash/v2"
)

type xxHasher struct{}

func (h *xxHasher) Hash(data []byte) uint64 {
	return xxhash.Sum64(data)
}
