module github.com/Grivn/phalanx

require (
	github.com/a8m/envsubst v1.1.0
	github.com/gogo/protobuf v1.2.1
	github.com/golang/mock v1.4.4
	github.com/golang/protobuf v1.4.2
	github.com/hyperledger-labs/minbft v0.0.0-20210107133144-8007d7dfea00
	github.com/op/go-logging v0.0.0-20160315200505-970db520ece7
	github.com/spf13/cobra v1.0.0
	github.com/spf13/viper v1.7.0
	github.com/stretchr/testify v1.6.1
	github.com/ultramesh/fancylogger v0.1.2
	github.com/ultramesh/flato-common v0.2.3
	golang.org/x/sync v0.0.0-20200317015054-43a5402ce75a
	google.golang.org/grpc v1.29.1
	google.golang.org/protobuf v1.24.0
	gopkg.in/yaml.v2 v2.3.0
)

replace github.com/ultramesh/crypto-standard => git.hyperchain.cn/ultramesh/crypto-standard.git v0.2.0

replace github.com/ultramesh/flato-common => git.hyperchain.cn/ultramesh/flato-common.git v0.2.3

replace github.com/ultramesh/fancylogger => git.hyperchain.cn/ultramesh/fancylogger.git v0.1.2

go 1.15
