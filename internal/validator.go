package internal

type Validator interface {
	Sign(payload []byte) ([][]byte, error)
	Verify(author uint64, signature [][]byte) error
}
