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
	ldkLogFilename = "ldk.log"
)

var Logger *logrus.Logger
var LDKLogger *logrus.Logger
var logFilePath string
var ldkLogFilePath string

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

func InitLDK(ldkLogLevel string) {
	LDKLogger = logrus.New()
	LDKLogger.SetFormatter(&logrus.JSONFormatter{})
	LDKLogger.SetOutput(os.Stdout)
	ldkLogrusLogLevel, err := strconv.Atoi(ldkLogLevel)
	if err != nil {
		ldkLogrusLogLevel = int(logrus.InfoLevel)
	}
	LDKLogger.SetLevel(logrus.Level(ldkLogrusLogLevel))
	if ldkLogrusLogLevel >= int(logrus.DebugLevel) {
		LDKLogger.ReportCaller = true
		LDKLogger.Debug("LDK Logrus report caller enabled in debug mode")
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

func AddLDKFileLogger(workdir string) error {
	ldkLogFilePath = filepath.Join(workdir, logDir, ldkLogFilename)
	ldkFileLoggerHook, err := lumberjackrus.NewHook(
		&lumberjackrus.LogFile{
			Filename:   ldkLogFilePath,
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
	LDKLogger.AddHook(ldkFileLoggerHook)
	return nil
}

func GetLogFilePath() string {
	return logFilePath
}

func GetLDKLogFilePath() string {
	return ldkLogFilePath
}
