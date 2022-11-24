package fetcher

import (
	"github.com/Grivn/phalanx/common/api"
	"github.com/Grivn/phalanx/common/protos"
	"github.com/Grivn/phalanx/external"
)

type checkpointFetcher struct {
	author uint64
	timer  api.LocalTimer

	// crypto is used to generate/verify certificates.
	crypto api.Crypto

	sender external.NetworkService
	logger external.Logger
}

func (cf *checkpointFetcher) ProcessCheckpointRequest(request *protos.CheckpointRequest) {}

func (cf *checkpointFetcher) ProcessCheckpointVote(vote *protos.CheckpointVote) {}

func (cf *checkpointFetcher) ProcessCheckpoint(checkpoint *protos.Checkpoint) {}
