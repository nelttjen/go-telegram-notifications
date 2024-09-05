package logger

import (
	"fmt"
	"github.com/sirupsen/logrus"
	logcfg "notifications-bot/pkg/logger/config"
)

func GetLogger(name string) (*Logger, error) {
	if !loggers.initialized {
		return nil, &NotInitializedError{}
	}

	logger, ok := loggers.loggers[name]
	var err error
	if !ok {
		err = &NoSuchLoggerError{}
	}

	return logger, err
}

func (l *Logger) Log(message string, level logrus.Level, requiredLevel int) {
	if !l.shouldLog(requiredLevel) {
		return
	}
	l.Logger.Log(level, message)
}

func (l *Logger) Logf(message string, level logrus.Level, requiredLevel int, args ...interface{}) {
	if !l.shouldLog(requiredLevel) {
		return
	}
	l.Logger.Logf(level, message, args...)
}

func (l *Logger) Logln(message string, level logrus.Level, requiredLevel int) {
	if !l.shouldLog(requiredLevel) {
		return
	}

	message = fmt.Sprintf("%s\n", message)

	l.Logger.Log(level, message)
}

func LogIfExists(loggerName string, message string, level logrus.Level, requiredLevel int) {
	logger, err := GetLogger(loggerName)
	if err != nil {
		return
	}

	logger.Log(message, level, requiredLevel)
}

func LogfIfExists(loggerName string, format string, level logrus.Level, requiredLevel int, args ...interface{}) {
	logger, err := GetLogger(loggerName)
	if err != nil {
		return
	}

	message := fmt.Sprintf(format, args...)

	logger.Log(message, level, requiredLevel)
}

func LoglnIfExists(loggerName string, message string, level logrus.Level, requiredLevel int) {
	logger, err := GetLogger(loggerName)
	if err != nil {
		return
	}

	message = fmt.Sprintf("%s\n", message)

	logger.Log(message, level, requiredLevel)
}

func LogflnIfExists(loggerName string, format string, level logrus.Level, requiredLevel int, args ...interface{}) {
	format = fmt.Sprintf("%s\n", format)
	LogfIfExists(loggerName, format, level, requiredLevel, args...)
}

func FastTrace(loggerName string, format string, args ...interface{}) {
	LogflnIfExists(loggerName, format, logrus.TraceLevel, logcfg.LoggerLevelTrace, args...)
}

func FastDebug(loggerName string, message string, args ...interface{}) {
	LogflnIfExists(loggerName, message, logrus.DebugLevel, logcfg.LoggerLevelDebug, args...)
}

func FastInfo(loggerName string, message string, args ...interface{}) {
	LogflnIfExists(loggerName, message, logrus.InfoLevel, logcfg.LoggerLevelImportant, args...)
}

func FastWarn(loggerName string, message string, args ...interface{}) {
	LogflnIfExists(loggerName, message, logrus.WarnLevel, logcfg.LoggerLevelImportant, args...)
}

func FastError(loggerName string, message string, args ...interface{}) {
	LogflnIfExists(loggerName, message, logrus.ErrorLevel, logcfg.LoggerLevelImportant, args...)
}

func FastFatal(loggerName string, message string, args ...interface{}) {
	LogflnIfExists(loggerName, message, logrus.FatalLevel, logcfg.LoggerLevelImportant, args...)
}
