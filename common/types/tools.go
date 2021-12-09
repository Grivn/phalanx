package types

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"time"
)

// MetricsInfo tracks the metrics info of phalanx.
type MetricsInfo struct {
	// AvePackOrderLatency indicates interval since receive command to generate pre-order.
	AvePackOrderLatency float64

	// AveOrderLatency indicates interval since receive command to generate partial order.
	AveOrderLatency float64

	// AveLogLatency indicates interval since generate partial order to commit partial order.
	AveLogLatency float64

	// AveCommandInfoLatency indicates since generate command info to commit command.
	AveCommandInfoLatency float64

	// SafeCommandCount indicates the number of command committed from safe path.
	SafeCommandCount int

	// RiskCommandCount indicates the number of command committed from risk path.
	RiskCommandCount int
}

//============================= Calculate Basic Information ===========================================

// CalculateFault calculates the upper fault amount in byzantine system with n nodes.
func CalculateFault(n int) int {
	return (n-1)/3
}

// CalculateQuorum calculates the quorum legal committee for byzantine system.
func CalculateQuorum(n int) int {
	return n-CalculateFault(n)
}

// CalculateOneCorrect calculates the lowest amount for set with at least one trusted node in byzantine system.
func CalculateOneCorrect(n int) int {
	return CalculateFault(n)+1
}

//==================================== Time Convert =============================================

func NanoToSecond(nano int64) float64 {
	return float64(nano)/float64(time.Second)
}

func NanoToMillisecond(nano int64) float64 {
	return float64(nano)/float64(time.Millisecond)
}

//================================== Struct Convert =======================================

func BytesToString(b []byte) string {
	return hex.EncodeToString(b)
}

func StringToBytes(str string) []byte {
	b, err := hex.DecodeString(str)
	if err != nil {
		panic(err)
	}
	return b
}

func Uint64MapToList(m map[uint64]bool) []uint64 {
	var list []uint64
	for key := range m {
		list = append(list, key)
	}
	return list
}

func Uint64ToBytes(num uint64) []byte {
	buffer := bytes.NewBuffer([]byte{})

	err := binary.Write(buffer, binary.BigEndian, &num)
	if err != nil {
		panic("error convert bytes to int")
	}

	return buffer.Bytes()
}

func BytesToUint64(bys []byte) uint64 {
	buffer := bytes.NewBuffer([]byte{})

	var num uint64
	err := binary.Read(buffer, binary.BigEndian, &num)
	if err != nil {
		panic("error convert bytes to int")
	}

	return num
}
