package usecase

type Hasher interface {
	Hash(plain string) (string, error)
	Compare(hash string, plain string) error
}
