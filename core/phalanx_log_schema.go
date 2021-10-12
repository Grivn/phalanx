package phalanx

import (
	"github.com/Grivn/phalanx/common/types"
	"github.com/Grivn/phalanx/external"
	"os"
	"strconv"
)

type moduleLogger struct {
	logManagerLog   external.Logger
	executorLog     external.Logger
	testLog         external.Logger
}

func newPLogger(logger external.Logger, divided bool, author uint64) (*moduleLogger, error) {
	// print phalanx logs in system file.
	if !divided {
		return &moduleLogger{
			logManagerLog:   logger,
			executorLog:     logger,
			testLog:         logger,
		}, nil
	}

	// print phalanx logs in divided files.
	logDir := "phalanx_node"+strconv.Itoa(int(author))
	_ = os.Mkdir(logDir, os.ModePerm)
	//if err != nil {
	//	logger.Errorf("Mkdir Failed: %s", err)
	//	return nil, err
	//}

	return &moduleLogger{
		logManagerLog:   types.NewRawLoggerFile(logDir+"/log-manager"),
		executorLog:     types.NewRawLoggerFile(logDir+"/executor"),
		testLog:         types.NewRawLoggerFile(logDir+"/test-client"),
	}, nil
}
