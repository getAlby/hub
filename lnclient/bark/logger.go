//go:build (darwin && (amd64 || arm64)) || (linux && (amd64 || arm64)) || (windows && amd64)

package bark

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/orandin/lumberjackrus"
	"github.com/sirupsen/logrus"
	bark "gitlab.com/ark-bitcoin/bark-ffi-bindings/golang/bark"

	"github.com/getAlby/hub/logger"
)

const barkLogFilename = "bark.log"
const logDir = "log"

// setLoggerOnce guards bark.SetLogger, which can only be installed once per
// process (calling it again returns an error from the underlying bridge).
var setLoggerOnce sync.Once

// barkLogger forwards bark's internal log records to a dedicated logrus logger
// (separate from the main app logger, with its own level and log file) so bark
// logs can be tuned independently of the rest of the hub.
type barkLogger struct {
	logger *logrus.Logger
}

var _ bark.BarkLogger = &barkLogger{}

// Log implements bark.BarkLogger. The record's bark log level is mapped onto the
// equivalent logrus level, and the bark target is attached as a structured field.
//
// IMPORTANT: this must not call back into any bark API that itself emits log
// records, or the foreign runtime may stack-overflow or deadlock.
func (l *barkLogger) Log(level bark.LogLevel, target string, message string) {
	l.logger.WithFields(logrus.Fields{
		"log_type": "bark",
		"target":   target,
	}).Log(barkLevelToLogrus(level), message)
}

func barkLevelToLogrus(level bark.LogLevel) logrus.Level {
	switch level {
	case bark.LogLevelError:
		return logrus.ErrorLevel
	case bark.LogLevelWarn:
		return logrus.WarnLevel
	case bark.LogLevelInfo:
		return logrus.InfoLevel
	case bark.LogLevelDebug:
		return logrus.DebugLevel
	case bark.LogLevelTrace:
		return logrus.TraceLevel
	}
	logger.Logger.WithField("log_level", level).Error("Unknown Bark log level")
	return logrus.ErrorLevel
}

// logrusToBarkLevel maps a logrus level onto the bark log level used to cap
// what the FFI bridge emits, so filtered-out records don't cross the boundary.
func logrusToBarkLevel(level logrus.Level) bark.LogLevel {
	switch level {
	case logrus.PanicLevel, logrus.FatalLevel, logrus.ErrorLevel:
		return bark.LogLevelError
	case logrus.WarnLevel:
		return bark.LogLevelWarn
	case logrus.InfoLevel:
		return bark.LogLevelInfo
	case logrus.DebugLevel:
		return bark.LogLevelDebug
	default:
		return bark.LogLevelTrace
	}
}

// installBarkLogger wires bark's internal logs into a dedicated logrus logger.
// It is safe to call more than once (e.g. when reopening the wallet); only the
// first call installs the logger, as the bridge can only be set once per
// process.
func installBarkLogger(logLevel logrus.Level, logToFile bool, workDir string) {
	setLoggerOnce.Do(func() {
		barkLogrus, err := createBarkLogger(logLevel, logToFile, workDir)
		if err != nil {
			logger.Logger.WithError(err).Error("Failed to create Bark logger")
			return
		}
		// Cap the FFI bridge at the configured level so records below it are
		// never produced; the dedicated logger applies the same level again.
		if err := bark.SetLogger(&barkLogger{logger: barkLogrus}, logrusToBarkLevel(logLevel)); err != nil {
			logger.Logger.WithError(err).Warn("Failed to install Bark logger")
		}
	})
}

func createBarkLogger(logLevel logrus.Level, logToFile bool, workDir string) (*logrus.Logger, error) {
	barkLogrus := logrus.New()
	barkLogrus.SetFormatter(&logrus.JSONFormatter{})
	barkLogrus.SetOutput(os.Stdout)
	barkLogrus.SetLevel(logLevel)

	if logToFile {
		parentDir := filepath.Dir(workDir)
		barkLogFilePath := filepath.Join(parentDir, logDir, barkLogFilename)
		barkFileLoggerHook, err := lumberjackrus.NewHook(
			&lumberjackrus.LogFile{
				Filename:   barkLogFilePath,
				MaxAge:     3,
				MaxBackups: 3,
			},
			logLevel,
			&logrus.JSONFormatter{},
			nil,
		)
		if err != nil {
			return nil, err
		}
		barkLogrus.AddHook(barkFileLoggerHook)
	}
	return barkLogrus, nil
}
