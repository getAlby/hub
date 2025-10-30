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
		}).Log(logger.MapLdkLogLevel(record.Level), record.Args)
	}
}
