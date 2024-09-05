package logger

import (
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	easy "github.com/t-tomalak/logrus-easy-formatter"
	"io"
	logcfg "notifications-bot/pkg/logger/config"
	"os"
)

type NotInitializedError struct{}

func (e NotInitializedError) Error() string {
	return "logger not initialized"
}

type NoSuchLoggerError struct{}

func (e NoSuchLoggerError) Error() string {
	return "no such logger initialized with this name"
}

type Logger struct {
	Logger *logrus.Logger
	Config logcfg.LoggingConfig
}

type loggerContainer struct {
	loggers     map[string]*Logger
	initialized bool
}

var loggers = &loggerContainer{initialized: false}

func (l *Logger) shouldLog(requiredLevel int) bool {
	if l.Config.GetLoggerEnabled() == logcfg.LogDisabled {
		return false
	}

	if l.Config.GetLoggerLevel() < requiredLevel {
		return false
	}

	if l.Logger == nil {
		return false
	}

	return true
}

func InitLoggers(cfg logcfg.LoggingConfig) {
	loggers = &loggerContainer{loggers: make(map[string]*Logger), initialized: true}
	count := 0

	if cfg.GetLoggerEnabled() == logcfg.LogDisabled {
		return
	}

	settingsFile, err := os.Open(cfg.GetAppRoot() + "/pkg/logger/config/settings.json")
	if err != nil {
		panic(fmt.Sprintf("Error opening settings logger file: %v", err))
	}
	defer func() {
		err := settingsFile.Close()
		if err != nil {
			FastWarn(logcfg.InfoLoggerName, "Error closing settings logger file: %v", err)
		}
	}()

	var settings map[string]interface{}

	bytes, _ := io.ReadAll(settingsFile)

	err = json.Unmarshal(bytes, &settings)
	if err != nil {
		panic(fmt.Sprintf("Error parsing settings logger file: %v", err))
	}

	var loggersKey string
	var formatter interface{}
	if !cfg.IsDevEnvironment() {
		loggersKey = "init_loggers_prod"
		formatter = &logrus.JSONFormatter{}
	} else {
		loggersKey = "init_loggers_dev"
		formatter = &easy.Formatter{
			TimestampFormat: "2006-01-02 15:04:05",
			LogFormat:       "[%lvl%]: %time% - %msg%",
		}
	}

	err = os.MkdirAll(cfg.GetAppRoot()+"/logs", 0777)
	if err != nil {
		panic(err)
	}

	for _, key := range settings[loggersKey].([]interface{}) {
		loggerSettings := settings["loggers"].(map[string]interface{})[key.(string)].(map[string]interface{})

		var level logrus.Level

		switch loggerSettings["level"].(string) {
		case "TRACE":
			level = logrus.TraceLevel
		case "DEBUG":
			level = logrus.DebugLevel
		case "INFO":
			level = logrus.InfoLevel
		case "WARN":
			level = logrus.WarnLevel
		case "ERROR":
			level = logrus.ErrorLevel
		default:
			level = logrus.InfoLevel
		}

		var out io.Writer
		var outs []interface{}

		handlers := loggerSettings["handlers"].([]interface{})

		for _, handler := range handlers {
			handler := settings["handlers"].(map[string]interface{})[handler.(string)].(map[string]interface{})
			switch handler["out"] {
			case "stdout":
				outs = append(outs, os.Stdout)
			case "file":
				file, err := os.OpenFile(cfg.GetAppRoot()+fmt.Sprintf("/logs/%s", handler["filename"].(string)), os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
				if err != nil {
					panic(err)
				}
				outs = append(outs, file)
			}
		}
		if len(outs) == 1 {
			out = outs[0].(io.Writer)
		} else {
			var multiOuts []io.Writer
			for _, o := range outs {
				multiOuts = append(multiOuts, o.(io.Writer))
			}
			out = io.MultiWriter(multiOuts...).(io.Writer)
		}

		logger := &logrus.Logger{
			Level:     level,
			Formatter: formatter.(logrus.Formatter),
			Out:       out,
		}
		loggers.loggers[key.(string)] = &Logger{Logger: logger, Config: cfg}
		count += 1
	}

	LogflnIfExists("info", "%d Loggers initialized", logrus.InfoLevel, logcfg.LoggerLevelImportant, count)
}
