package internal

type Executor interface {
	// CommitQCs is used to commit the QCs.
	CommitQCs(payload []byte) error
}
