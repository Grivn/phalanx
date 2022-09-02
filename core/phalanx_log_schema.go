package phalanx

import (
	"os"
	"strconv"
	
	"github.com/Grivn/phalanx/common/types"
	"github.com/Grivn/phalanx/external"
)

type moduleLogger struct {
	metaPoolLog  external.Logger
	executorLog  external.Logger
	txManagerLog external.Logger
}

func newPLogger(logger external.Logger, divided bool, author uint64) (*moduleLogger, error) {
	// print phalanx logs in system file.
	if !divided {
		return &moduleLogger{
			metaPoolLog:  logger,
			executorLog:  logger,
			txManagerLog: logger,
		}, nil
	}

	// print phalanx logs in divided files.
	logDir := "phalanx_node" + strconv.Itoa(int(author))
	err := os.RemoveAll(logDir)
	if err != nil {
		logger.Errorf("Mkdir Failed: %s", err)
		return nil, err
	}
	err = os.Mkdir(logDir, os.ModePerm)
	if err != nil {
		logger.Errorf("Mkdir Failed: %s", err)
		return nil, err
	}

	return &moduleLogger{
		metaPoolLog:  types.NewRawLoggerFile(logDir + "/log-manager"),
		executorLog:  types.NewRawLoggerFile(logDir + "/executor"),
		txManagerLog: types.NewRawLoggerFile(logDir + "/test-client"),
	}, nil
}
