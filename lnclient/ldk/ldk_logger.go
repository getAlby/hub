package ldk

import (
	"os"
	"path/filepath"

	"github.com/getAlby/hub/logger"
	"github.com/getAlby/ldk-node-go/ldk_node"
	"github.com/orandin/lumberjackrus"
	"github.com/sirupsen/logrus"
)

const ldkLogFilename = "ldk.log"
const logDir = "log"

type ldkLogger struct {
	logLevel logrus.Level
	logger   *logrus.Logger
}

func NewLDKLogger(logLevel logrus.Level, logToFile bool, workDir string) (ldk_node.LogWriter, error) {
	logger, err := createLogger(logLevel, logToFile, workDir)
	if err != nil {
		return nil, err
	}
	return &ldkLogger{
		logLevel: logLevel,
		logger:   logger,
	}, nil
}

func (ldkLogger *ldkLogger) Log(record ldk_node.LogRecord) {
	ldkLogger.logger.WithFields(logrus.Fields{
		"log_type":    "LDK-node",
		"line":        record.Line,
		"module_path": record.ModulePath,
	}).Log(mapLogLevel(record.Level), record.Args)
}

func mapLogLevel(logLevel ldk_node.LogLevel) logrus.Level {
	switch logLevel {
	case ldk_node.LogLevelGossip:
		return logrus.TraceLevel
	case ldk_node.LogLevelTrace:
		return logrus.TraceLevel
	case ldk_node.LogLevelDebug:
		return logrus.DebugLevel
	case ldk_node.LogLevelInfo:
		return logrus.InfoLevel
	case ldk_node.LogLevelWarn:
		return logrus.WarnLevel
	case ldk_node.LogLevelError:
		return logrus.ErrorLevel
	}
	logger.Logger.WithField("log_level", logLevel).Error("Unknown LDK log level")
	return logrus.ErrorLevel
}

func createLogger(logLevel logrus.Level, logToFile bool, workDir string) (*logrus.Logger, error) {
	ldkLogger := logrus.New()
	ldkLogger.SetFormatter(&logrus.JSONFormatter{})
	ldkLogger.SetOutput(os.Stdout)

	ldkLogger.SetLevel(logLevel)

	if logToFile {
		parentDir := filepath.Dir(workDir)
		ldkLogFilePath := filepath.Join(parentDir, logDir, ldkLogFilename)
		ldkFileLoggerHook, err := lumberjackrus.NewHook(
			&lumberjackrus.LogFile{
				Filename:   ldkLogFilePath,
				MaxAge:     3,
				MaxBackups: 3,
			},
			logLevel,
			&logrus.JSONFormatter{},
			nil,
		)
		if err != nil {
			logger.Logger.WithError(err).Error("Failed to add LDK file logger")
			return nil, err
		}

		ldkLogger.AddHook(ldkFileLoggerHook)
	}
	return ldkLogger, nil
}
