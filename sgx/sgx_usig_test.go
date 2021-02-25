package sgx

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	enclaveFile = "enclave/libusig.signed.so"
)

func TestSGXUSIG(t *testing.T) {
	msg := []byte("Test message")
	wrongMsg := []byte("Another message")

	// Create the first USIG instance to generate a new USIG key
	usig, err := NewUSIG(enclaveFile, nil)
	assert.NoError(t, err, "Error creating fist SGXUSIG instance")
	assert.NotNil(t, usig, "Got nil SGXUSIG instance")

	// Get the sealed key from the first instance
	key := usig.enclave.SealedKey()
	assert.NotNil(t, key, "Got nil sealed key")

	// Get the public key from the fist instance
	usigPubKey := usig.enclave.PublicKey()
	assert.NotNil(t, usigPubKey, "Got nil USIG public key")

	// Destroy the fist instance, just to be sure
	usig.enclave.Destroy()

	// Recreate USIG restoring the key from the first instance
	usig, err = NewUSIG(enclaveFile, key)
	assert.NoError(t, err, "Error creating SGXUSIG instance with key unsealing")
	assert.NotNil(t, usig, "Got nil SGXUSIG instance")
	defer usig.enclave.Destroy()

	epoch := usig.enclave.Epoch()
	usigID, err := MakeID(epoch, usigPubKey)
	require.NoError(t, err)

	parsedEpoch, parsedPubKey, err := ParseID(usigID)
	assert.NoError(t, err)
	assert.Equal(t, epoch, parsedEpoch)
	assert.Equal(t, usigPubKey, parsedPubKey)

	ui, err := usig.CreateUI(msg)
	assert.NoError(t, err, "Error creating UI")
	assert.NotNil(t, ui, "Got nil UI")
	assert.Equal(t, uint64(1), ui.Counter, "Got wrong UI counter value")

	ui, err = usig.CreateUI(msg)
	assert.NoError(t, err, "Error creating UI")
	assert.NotNil(t, ui, "Got nil UI")
	assert.Equal(t, uint64(2), ui.Counter, "Got wrong UI counter value")

	// There's no need to repeat all the checks covered by C test
	// of the enclave. But correctness of the signature should be
	// checked.
	err = VerifyUI(msg, ui, usigID)
	assert.NoError(t, err, "Error verifying UI")

	err = VerifyUI(wrongMsg, ui, usigID)
	assert.Error(t, err, "No error verifying UI with forged message")
}
