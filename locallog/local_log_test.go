package locallog

import (
	"github.com/Grivn/phalanx/common/crypto"
	"strconv"
	"testing"

	"github.com/Grivn/phalanx/common/mocks"
	"github.com/Grivn/phalanx/common/protos"

	"github.com/stretchr/testify/assert"
)

func TestLocalLog(t *testing.T) {
	n := 4
	lls := make(map[uint64]*localLog)
	logger := mocks.NewRawLogger()

	_ = crypto.SetKeys()

	for i:=0; i<n; i++ {
		lls[uint64(i+1)] = NewLocalLog(n, uint64(i+1), logger)
	}

	count := 100
	for i:=0; i<count; i++ {
		c := &protos.Command{Digest: "test-digest"+strconv.Itoa(i)}
		pre, _ := lls[1].ProcessCommand(c)
		assert.Equal(t, uint64(i+1), pre.Sequence)
	}
}
