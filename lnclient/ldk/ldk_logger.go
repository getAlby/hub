package ldk

import (
	// "github.com/getAlby/hub/ldk_node"
	"github.com/getAlby/hub/logger"

	"github.com/getAlby/ldk-node-go/ldk_node"
	"github.com/sirupsen/logrus"
)

type ldkLogger struct {
	logLevel ldk_node.LogLevel
}

func NewLDKLogger(logLevel ldk_node.LogLevel) ldk_node.LogWriter {
	return &ldkLogger{
		logLevel: logLevel,
	}
}

func (ldkLogger *ldkLogger) Log(record ldk_node.LogRecord) {
	if record.Level >= ldkLogger.logLevel {
		logger.LDKLogger.WithFields(logrus.Fields{
			"log_type":    "LDK-node",
			"line":        record.Line,
			"module_path": record.ModulePath,
		}).Log(mapLogLevel(record.Level), record.Args)
	}
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
	logger.LDKLogger.WithField("log_level", logLevel).Error("Unknown LDK log level")
	return logrus.ErrorLevel
}
