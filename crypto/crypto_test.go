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

	cert, err := PrivSign(hash, 1)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(cert.Signatures))

	flag, err := PubVerify(cert, hash, 1)
	assert.True(t, flag)
	assert.Nil(t, err)
}
