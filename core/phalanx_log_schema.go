package phalanx

import (
	"os"
	"strconv"
	"time"

	"github.com/Grivn/phalanx/common/types"
	"github.com/Grivn/phalanx/external"
)

type moduleLogger struct {
	logManagerLog   external.Logger
	sequencePoolLog external.Logger
	executorLog     external.Logger
}

func newPLogger(logger external.Logger, divided bool, author uint64) (*moduleLogger, error) {
	// print phalanx logs in system file.
	if !divided {
		return &moduleLogger{
			logManagerLog:   logger,
			sequencePoolLog: logger,
			executorLog:     logger,
		}, nil
	}

	// print phalanx logs in divided files.
	logDir := "phalanx_node"+strconv.Itoa(int(author))+"_"+time.Now().Format("2006-01-02_15:04:05")
	err := os.Mkdir(logDir, os.ModePerm)
	if err != nil {
		logger.Errorf("Mkdir Failed: %s", err)
		return nil, err
	}

	return &moduleLogger{
		logManagerLog:   types.NewRawLoggerFile(logDir+"/log-manager"),
		sequencePoolLog: types.NewRawLoggerFile(logDir+"/sequence-pool"),
		executorLog:     types.NewRawLoggerFile(logDir+"/executor"),
	}, nil
}
