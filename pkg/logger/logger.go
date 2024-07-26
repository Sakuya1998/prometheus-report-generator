package logger

import (
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

func InitLogger(path string, maxAge int, compress bool) map[logrus.Level]*logrus.Logger {
	loggers := make(map[logrus.Level]*logrus.Logger)

	for _, level := range []logrus.Level{logrus.DebugLevel, logrus.InfoLevel, logrus.WarnLevel, logrus.ErrorLevel} {
		log := logrus.New()
		log.SetLevel(level)
		log.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
		})

		// Create the directory if it doesn't exist
		if _, err := os.Stat(path); os.IsNotExist(err) {
			os.MkdirAll(path, 0755)
		}

		logFile := &lumberjack.Logger{
			Filename:  filepath.Join(path, level.String()+".log"),
			MaxAge:    maxAge,   // days
			Compress:  compress, // whether to compress
			LocalTime: true,
		}

		log.SetOutput(logFile)
		log.SetReportCaller(true)

		loggers[level] = log
	}

	return loggers
}

func Log(loggers map[logrus.Level]*logrus.Logger, level logrus.Level, message string) {
	if logger, ok := loggers[level]; ok {
		logger.Log(level, message)
	}
}
