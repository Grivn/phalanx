package types

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"

	nested "github.com/antonfisher/nested-logrus-formatter"
	"github.com/sirupsen/logrus"
)

// CalculateFault calculates the upper fault amount in byzantine system with n nodes.
func CalculateFault(n int) int {
	return (n-1)/3
}

// CalculateQuorum calculates the quorum legal committee for byzantine system.
func CalculateQuorum(n int) int {
	return n-CalculateFault(n)
}

// CalculateOneCorrect calculates the lowest amount for set with at least one trusted node in byzantine system.
func CalculateOneCorrect(n int) int {
	return CalculateFault(n)+1
}

// NewRawLogger provides a Logger instance to print logs in console.
func NewRawLogger() *logrus.Logger {
	log := logrus.New()
	writer := os.Stdout
	log.SetLevel(logrus.DebugLevel)
	log.SetFormatter(formatter(true))
	log.SetOutput(io.MultiWriter(writer))
	return log
}

// NewRawLoggerFile provides a Logger instance to print logs in files.
func NewRawLoggerFile(path string) *logrus.Logger {
	log := logrus.New()
	writer, err := os.OpenFile(path+".log", os.O_WRONLY|os.O_CREATE, os.ModePerm)
	if err != nil {
		panic(err)
	}
	log.SetLevel(logrus.DebugLevel)
	log.SetFormatter(formatter(false))
	log.SetOutput(io.MultiWriter(writer))
	return log
}

// formatter is used to generate log format.
func formatter(isConsole bool) *nested.Formatter {
	// custom function to generate filename and line.
	customFunc := func(frame *runtime.Frame) string {
		funcInfo := runtime.FuncForPC(frame.PC)
		if funcInfo == nil {
			return "formatter error: FuncForPC failed"
		}
		fullPath, line := funcInfo.FileLine(frame.PC)
		return fmt.Sprintf(" [%v:%v]", filepath.Base(fullPath), line)
	}

	// generate formatter.
	format := &nested.Formatter{
		HideKeys:              true,
		TimestampFormat:       "2006-01-02 15:04:05",
		CallerFirst:           true,
		CustomCallerFormatter: customFunc,
	}

	// generate color in console.
	if isConsole {
		format.NoColors = false
	} else {
		format.NoColors = true
	}

	return format
}
