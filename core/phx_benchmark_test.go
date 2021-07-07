package phalanx

import (
	"testing"

	"github.com/Grivn/phalanx/common/crypto"
	"github.com/Grivn/phalanx/common/types"
)

// Benchmark test for QCBatch marshal/unmarshal

var count = 500
var size = 100

func BenchmarkNewPhalanxProvider3(b *testing.B) {
	_ = crypto.SetKeys()
	command := types.GenerateRandCommand(1, 1, count, size)
	b.ResetTimer()
	for i:=0; i<b.N; i++ {
		_, _ = crypto.PrivSign(types.StringToBytes(command.Digest), 1)
	}
	b.StopTimer()
}

func BenchmarkNewPhalanxProvider4(b *testing.B) {
	_ = crypto.SetKeys()
	command := types.GenerateRandCommand(1, 1, count, size)
	sig, _ := crypto.PrivSign(types.StringToBytes(command.Digest), 1)
	b.ResetTimer()
	for i:=0; i<b.N; i++ {
		for i:=0; i<types.COUNT; i++ {
			for i:=0; i<types.COUNT; i++ {
				_ = crypto.PubVerify(sig, types.StringToBytes(command.Digest), 1)
			}
		}
	}
	b.StopTimer()
}
