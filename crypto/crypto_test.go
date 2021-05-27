package crypto

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCrypto(t *testing.T) {
	err := SetKeys()
	assert.Nil(t, err)

	payload := []byte("test message")
	id := MakeID(payload)
	hash := IDToByte(id)

	sig, err := PrivSign(hash, 1)
	assert.Nil(t, err)

	flag, err := PubVerify(sig, hash, 1)
	assert.True(t, flag)
	assert.Nil(t, err)
}
