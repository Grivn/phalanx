//go:generate mockgen -destination ../common/mocks/mock_network.go -package mocks -source network.go

//go:generate mockgen -destination ../common/mocks/mock_executor.go -package mocks -source executor.go

//go:generate mockgen -destination ../common/mocks/mock_logger.go -package mocks -source logger.go

//go:generate mockgen -destination ../common/mocks/mock_crypto.go -package mocks -source crypto.go

package external
