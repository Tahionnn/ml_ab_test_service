package hasher

type Hasher interface {
	Hash(data []byte) uint64
}

func NewHasher() Hasher {
	return &xxHasher{}
}
