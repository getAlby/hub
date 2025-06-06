package logger

import (
	"os"
	"path/filepath"
	"strconv"

	"github.com/orandin/lumberjackrus"
	"github.com/sirupsen/logrus"
)

const (
	logDir      = "log"
	logFilename = "nwc.log"
)

var Logger *logrus.Logger
var logFilePath string

func Init(logLevel string) {
	Logger = logrus.New()
	Logger.SetFormatter(&logrus.JSONFormatter{})
	Logger.SetOutput(os.Stdout)
	logrusLogLevel, err := strconv.Atoi(logLevel)
	if err != nil {
		logrusLogLevel = int(logrus.InfoLevel)
	}
	Logger.SetLevel(logrus.Level(logrusLogLevel))
	if logrusLogLevel >= int(logrus.DebugLevel) {
		Logger.ReportCaller = true
		Logger.Debug("Logrus report caller enabled in debug mode")
	}
}

func AddFileLogger(workdir string) error {
	logFilePath = filepath.Join(workdir, logDir, logFilename)
	fileLoggerHook, err := lumberjackrus.NewHook(
		&lumberjackrus.LogFile{
			Filename:   logFilePath,
			MaxAge:     3,
			MaxBackups: 3,
		},
		logrus.InfoLevel,
		&logrus.JSONFormatter{},
		nil,
	)
	if err != nil {
		return err
	}
	Logger.AddHook(fileLoggerHook)
	return nil
}

func GetLogFilePath() string {
	return logFilePath
}
