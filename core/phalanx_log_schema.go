package phalanx

import (
	"io"
	"os"
	"strconv"
	"time"

	"github.com/Grivn/phalanx/external"

	nested "github.com/antonfisher/nested-logrus-formatter"
	"github.com/sirupsen/logrus"
)

type moduleLogger struct {
	logManagerLog   external.Logger
	sequencePoolLog external.Logger
	executorLog     external.Logger
}

func newPLogger(logger external.Logger, divided bool, author uint64) (*moduleLogger, error) {
	// print phalanx logs in one file.
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
		logManagerLog:   generateLogger(logDir+"/log-manager"),
		sequencePoolLog: generateLogger(logDir+"/sequence-pool"),
		executorLog:     generateLogger(logDir+"/executor"),
	}, nil
}

func generateLogger(name string) external.Logger {
	log := logrus.New()
	writer, err := os.OpenFile(name+".log", os.O_WRONLY|os.O_CREATE, 0755)
	if err != nil {
		panic(err)
	}
	log.SetLevel(logrus.DebugLevel)
	log.SetFormatter(&nested.Formatter{NoColors: true})
	log.SetOutput(io.MultiWriter(writer))
	return log
}
