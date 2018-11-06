package log

import (
	"io"
	"log"
	"os"
	"sync"
)

var (
	loggerMu sync.Mutex
	logger   *log.Logger
	verbose  bool
)

func init() {
	logger = new(log.Logger)
	logger.SetOutput(os.Stderr)
}

// SetVerbose sets whether or not verbose debug logs are outputted.
func SetVerbose(v bool) {
	loggerMu.Lock()
	defer loggerMu.Unlock()
	verbose = v
}

// SetOutput sets the log output
func SetOutput(out io.Writer) {
	loggerMu.Lock()
	defer loggerMu.Unlock()
	logger.SetOutput(out)
}

// Debug prints debug output
func Debug(v ...interface{}) {
	loggerMu.Lock()
	defer loggerMu.Unlock()
	if verbose {
		logger.Print(v...)
	}
}

// Debugf prints debugging output with formatting
func Debugf(format string, v ...interface{}) {
	loggerMu.Lock()
	defer loggerMu.Unlock()
	if verbose {
		logger.Printf(format, v...)
	}
}

// Info prints errors and updates that are always shown to the user
func Info(v ...interface{}) {
	loggerMu.Lock()
	defer loggerMu.Unlock()
	logger.Print(v...)
}

// Infof prints errors and updates always shown to the user, with formatting
func Infof(format string, v ...interface{}) {
	loggerMu.Lock()
	defer loggerMu.Unlock()
	logger.Printf(format, v...)
}
