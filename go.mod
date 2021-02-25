module github.com/Grivn/phalanx

require (
	github.com/gogo/protobuf v1.2.1
	github.com/golang/mock v1.4.4
	github.com/stretchr/testify v1.6.1
	github.com/ultramesh/fancylogger v0.1.2
	github.com/ultramesh/flato-common v0.2.3
)

replace github.com/ultramesh/crypto-standard => git.hyperchain.cn/ultramesh/crypto-standard.git v0.2.0

replace github.com/ultramesh/flato-common => git.hyperchain.cn/ultramesh/flato-common.git v0.2.3

replace github.com/ultramesh/fancylogger => git.hyperchain.cn/ultramesh/fancylogger.git v0.1.2

go 1.15
