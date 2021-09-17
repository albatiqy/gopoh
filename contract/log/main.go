package log

import (
	"io"
	"log"
	"os"
	"runtime"
	"strconv"
)

type writer struct {
}

func (w writer) Write(p []byte) (n int, err error) {
	prefix := []byte("INFO: ")
	prefix = append(prefix, p...)
	logger.Writer().Write(prefix)
	return len(prefix), nil
}

const (
	LevelFatal = iota
	LevelError
	LevelWarning
	LevelInfo
	LevelDebug
)

var (
	logger       *log.Logger
	currentLevel int
)

func init() { // kasih instance
	logger = log.New(os.Stdout, "", log.Ldate|log.Ltime)
	currentLevel = LevelDebug
}

func Writer() io.Writer {
	return writer{}
}

func Default() *log.Logger {
	return logger
}

func NewLogger(prefix string) *log.Logger {
	return log.New(os.Stdout, prefix, log.Ldate|log.Ltime)
}

func print(level int, prefix string, s ...interface{}) {
	if currentLevel >= level {
		logger.Print(s...)
	}
}

func printf(level int, format string, prefix string, s ...interface{}) {
	if currentLevel >= level {
		logger.Printf(format, s...)
	}
}

func SetLevel(level int) {
	currentLevel = level
}

func Print(s ...interface{}) {
	print(LevelInfo, "LOG: ", s...)
}

func Debug(s ...interface{}) { //buggy
	if currentLevel >= LevelDebug {
		if _, fileName, fileLine, ok := runtime.Caller(1); ok {
			logger.Printf("\nDEBUG:\n  message: \n    %s \n  in     : " + fileName + ":" + strconv.Itoa(fileLine) + "\n", s...)
		} else {
			logger.Print(s...)
		}
	}
}

func Debugf(format string, s ...interface{}) {
	printf(LevelDebug, format, "DEBUG: ", s...)
}

func Info(s ...interface{}) {
	print(LevelInfo, "INFO: ", s...)
}

func Infof(format string, s ...interface{}) {
	printf(LevelInfo, format, "INFO: ", s...)
}

func Warning(s ...interface{}) {
	print(LevelWarning, "WARNING: ", s...)
}

func Warningf(format string, s ...interface{}) {
	printf(LevelWarning, format, "WARNING: ", s...)
}

func Error(s ...interface{}) {
	print(LevelError, "ERROR: ", s...)
}

func Errorf(format string, s ...interface{}) {
	printf(LevelError, format, "ERROR: ", s...)
}

func Fatal(s ...interface{}) {
	print(LevelFatal, "FATAL: ", s...)
	os.Exit(1)
}

func Fatalf(format string, s ...interface{}) {
	printf(LevelFatal, format, "FATAL: ", s...)
	os.Exit(1)
}
