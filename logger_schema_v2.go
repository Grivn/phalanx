package phalanx

import (
	"os"
	"strconv"

	"github.com/Grivn/phalanx/pkg/common/types"
	"github.com/Grivn/phalanx/pkg/external"
)

type moduleLoggerV2 struct {
	proposerLog         external.Logger
	consensusEngineLog  external.Logger
	sequencingEngineLog external.Logger
	finalityEngine      external.Logger
	sequencerLog        external.Logger
	memoryPoolLog       external.Logger
	finalityLog         external.Logger
}

func newPLoggerV2(logger external.Logger, divided bool, author uint64) (*moduleLoggerV2, error) {
	// print phalanx logs in system file.
	if !divided {
		return &moduleLoggerV2{
			proposerLog:         logger,
			consensusEngineLog:  logger,
			sequencingEngineLog: logger,
			finalityEngine:      logger,
			sequencerLog:        logger,
			memoryPoolLog:       logger,
			finalityLog:         logger,
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

	return &moduleLoggerV2{
		proposerLog:         types.NewRawLoggerFile(logDir + "/proposerLog"),
		consensusEngineLog:  types.NewRawLoggerFile(logDir + "/consensusEngineLog"),
		sequencingEngineLog: types.NewRawLoggerFile(logDir + "/sequencingEngineLog"),
		finalityEngine:      types.NewRawLoggerFile(logDir + "/finalityEngine"),
		sequencerLog:        types.NewRawLoggerFile(logDir + "/sequencerLog"),
		memoryPoolLog:       types.NewRawLoggerFile(logDir + "/memoryPoolLog"),
		finalityLog:         types.NewRawLoggerFile(logDir + "/finalityLog"),
	}, nil
}
