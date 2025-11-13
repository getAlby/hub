package ldk

import (
	"os"
	"path/filepath"
	"strconv"

	"github.com/getAlby/hub/logger"
	"github.com/getAlby/ldk-node-go/ldk_node"
	"github.com/orandin/lumberjackrus"
	"github.com/sirupsen/logrus"
)

const ldkLogFilename = "ldk.log"
const logDir = "log"

var LDKLogger *logrus.Logger
var ldkLogFilePath string

type ldkLogger struct {
	logLevel ldk_node.LogLevel
}

func NewLDKLogger(logLevel ldk_node.LogLevel) ldk_node.LogWriter {
	return &ldkLogger{
		logLevel: logLevel,
	}
}

func (ldkLogger *ldkLogger) Log(record ldk_node.LogRecord) {
	LDKLogger.WithFields(logrus.Fields{
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

func InitLogger(ldkLogLevel string) {
	LDKLogger = logrus.New()
	LDKLogger.SetFormatter(&logrus.JSONFormatter{})
	LDKLogger.SetOutput(os.Stdout)
	ldkLogrusLogLevel, err := strconv.Atoi(ldkLogLevel)
	if err != nil {
		ldkLogrusLogLevel = int(logrus.InfoLevel)
	}
	LDKLogger.SetLevel(logrus.Level(ldkLogrusLogLevel))

	if int(mapLogLevel(ldk_node.LogLevel(ldkLogrusLogLevel))) >= int(logrus.DebugLevel) {
		LDKLogger.ReportCaller = true
		LDKLogger.Debug("LDK Logrus report caller enabled in debug mode")
	}
}

func AddFileLogger(workdir string) error {
	ldkLogFilePath = filepath.Join(workdir, logDir, ldkLogFilename)
	ldkFileLoggerHook, err := lumberjackrus.NewHook(
		&lumberjackrus.LogFile{
			Filename:   ldkLogFilePath,
			MaxAge:     3,
			MaxBackups: 3,
		},
		LDKLogger.Level,
		&logrus.JSONFormatter{},
		nil,
	)
	if err != nil {
		return err
	}
	LDKLogger.AddHook(ldkFileLoggerHook)
	return nil
}
