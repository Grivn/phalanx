package logmgr

import (
	"fmt"
	"github.com/Grivn/phalanx/api"
	authen "github.com/Grivn/phalanx/authentication"
	"github.com/a8m/envsubst"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestNewLogManager(t *testing.T) {
	id := uint32(viper.GetInt("replica.id"))

	usigEnclaveFile, err := envsubst.String(viper.GetString("usig.enclaveFile"))
	assert.Error(t, err)

	keysFile, err := os.Open(viper.GetString("keys"))
	assert.Error(t, err)

	auth, err := authen.NewWithSGXUSIG([]api.AuthenticationRole{api.ReplicaAuthen, api.USIGAuthen}, id, keysFile, usigEnclaveFile)
	assert.Error(t, err)

	fmt.Println(auth)
}
