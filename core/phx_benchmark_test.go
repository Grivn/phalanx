package phalanx

import (
	"strconv"
	"testing"

	"github.com/Grivn/phalanx/common/crypto"
	"github.com/Grivn/phalanx/common/protos"
	"github.com/Grivn/phalanx/common/types"
)

// Benchmark test for QCBatch marshal/unmarshal

var count = 500
var size = 100

func QCBatchGeneration() *protos.QCBatch {
	command := types.GenerateRandCommand(count, size)

	agg := make(map[uint64]*protos.Certification)

	for i:=0; i<types.COUNT; i++ {
		signature, err := crypto.PrivSign(types.StringToBytes(command.Digest), i+1)
		if err != nil {
			panic(err)
		}
		agg[uint64(i)] = signature
	}

	qc := &protos.QuorumCert{ProofCerts: &protos.ProofCerts{Certs: agg}}

	var qcs []*protos.QuorumCert

	for i:=0; i<types.COUNT; i++ {
		qcs = append(qcs, qc)
	}

	filter := &protos.QCFilter{QCs: qcs}

	filters := []*protos.QCFilter{filter}

	commands := make(map[string]*protos.Command)
	for i:=0; i<types.COUNT; i++ {
		commands[strconv.Itoa(i)] = command
	}

	qcb := &protos.QCBatch{Filters: filters, Commands: commands}
	return qcb
}

func BenchmarkNewPhalanxProvider1(b *testing.B) {
	_ = crypto.SetKeys()
	qcb := QCBatchGeneration()
	b.ResetTimer()
	for i:=0; i<b.N; i++ {
		_, _ = marshal(qcb)
	}
	b.StopTimer()
}

func BenchmarkNewPhalanxProvider2(b *testing.B) {
	_ = crypto.SetKeys()
	qcb := QCBatchGeneration()
	payload, _ := marshal(qcb)
	b.ResetTimer()
	for i:=0; i<b.N; i++ {
		_, _ = unmarshal(payload)
	}
	b.StopTimer()
}

func BenchmarkNewPhalanxProvider3(b *testing.B) {
	_ = crypto.SetKeys()
	command := types.GenerateRandCommand(count, size)
	b.ResetTimer()
	for i:=0; i<b.N; i++ {
		_, _ = crypto.PrivSign(types.StringToBytes(command.Digest), 1)
	}
	b.StopTimer()
}

func BenchmarkNewPhalanxProvider4(b *testing.B) {
	_ = crypto.SetKeys()
	command := types.GenerateRandCommand(count, size)
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
