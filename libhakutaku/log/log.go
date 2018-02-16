// Copyright (C) 2018 Mikhail Masyagin

/*
Package log contains OOP-container for the "github.com/mgutz/logxi/v1" logger.
*/
package log

import (
	"io"
	"os"

	"github.com/mgutz/logxi/v1"
)

// Logger struct contains information about logger.
// If Logger.UseLog is true, Logger will write logs, else will not.
// Logger.Logger contains logger.
// Logger.File contains pointer to log-file (nil, if Logger.UseLog is false
// or Logger.Logger writes to stdout).
type Logger struct {
	UseLog bool
	Logger log.Logger
	File   *os.File
}

// NewLogger creates new logger.
// Config contains log file name ("" for no logging, stdout for stdout logging)
func NewLogger(config string) (logger Logger, err error) {
	switch config {
	case "stdout":
		logger.UseLog = true
		logger.Logger = log.NewLogger(log.NewConcurrentWriter(os.Stdout), "hakutaku_bot")
	case "":
	default:
		logger.UseLog = true

		var logFile *os.File
		if _, err = os.Stat(config); os.IsNotExist(err) {
			logFile, err = os.Create(config)
			if err != nil {
				return
			}
		} else {
			logFile, err = os.Open(config)
			if err != nil {
				return
			}
		}
		logger.Logger = log.NewLogger(log.NewConcurrentWriter(io.Writer(logFile)), "hakutaku_bot")
	}
	return
}

// CloseLogger closes log-file.
func (logger *Logger) CloseLogger() {
	if logger.File != nil {
		logger.File.Close()
	}
}

// Info logs important infotmation.
func (logger *Logger) Info(msg string, arg ...interface{}) {
	if logger.UseLog {
		logger.Logger.Info(msg, arg...)
	}
}

// Error logs errors.
func (logger *Logger) Error(msg string, arg ...interface{}) (err error) {
	if logger.UseLog {
		err = logger.Logger.Error(msg, arg...)
	}
	return
}
