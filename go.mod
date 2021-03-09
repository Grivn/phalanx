module github.com/Grivn/phalanx

require (
	github.com/a8m/envsubst v1.1.0
	github.com/gogo/protobuf v1.2.1
	github.com/golang/mock v1.4.4
	github.com/spf13/cobra v1.0.0
	github.com/spf13/viper v1.7.0
	github.com/stretchr/testify v1.6.1
	github.com/ultramesh/fancylogger v0.1.2
	gopkg.in/yaml.v2 v2.3.0
)

replace github.com/ultramesh/fancylogger => git.hyperchain.cn/ultramesh/fancylogger.git v0.1.2

go 1.15
