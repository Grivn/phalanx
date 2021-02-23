module github.com/Grivn/phalanx

require (
	github.com/Grivn/phalanx-common v0.0.0
	github.com/ultramesh/fancylogger v0.1.2
)

replace github.com/Grivn/phalanx-common => ../phalanx-common

replace github.com/ultramesh/crypto-standard => git.hyperchain.cn/ultramesh/crypto-standard.git v0.2.0

replace github.com/ultramesh/flato-common => git.hyperchain.cn/ultramesh/flato-common.git v0.2.3

replace github.com/ultramesh/fancylogger => git.hyperchain.cn/ultramesh/fancylogger.git v0.1.2

go 1.15
