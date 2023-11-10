package logger

import (
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	easy "github.com/t-tomalak/logrus-easy-formatter"
	"io"
	"os"

	appConfig "notifications-bot/internal/config"
)

type notInitializedError struct {
}

func (e *notInitializedError) Error() string {
	return "logger not initialized"
}

type noSuchLogger struct{}

func (e *noSuchLogger) Error() string {
	return "no such logger initialized with this name"
}

var loggers []*Logger
var initialized = false

type TemplateLogger struct {
	Name       string
	Out        interface{}
	Level      logrus.Level
	Formatters []logrus.Formatter
}

type Logger struct {
	Logger *logrus.Logger
	name   string
}

func (l *Logger) IsThisName(name string) bool {
	return l.name == name
}

func (l *Logger) shouldLog(requiredLevel uint8) bool {
	if LoggerEnabled == LogDisabled || appConfig.Environment == appConfig.Testing {
		return false
	}

	if LoggerLevel < requiredLevel {
		return false
	}

	if l.Logger == nil {
		return false
	}

	return true
}

func InitLoggers() {
	initialized = true

	if LoggerEnabled == LogDisabled || appConfig.Environment == appConfig.Testing {
		loggers = append(loggers, &Logger{Logger: logrus.New(), name: "disabled"})
		return
	}

	settingsFile, err := os.Open(appConfig.AppRoot + "/pkg/logger/settings.json")
	if err != nil {
		panic(fmt.Sprintf("Error opening settings logger file: %v", err))
	}
	defer settingsFile.Close()

	var settings map[string]interface{}

	bytes, _ := io.ReadAll(settingsFile)

	err = json.Unmarshal(bytes, &settings)
	if err != nil {
		panic(fmt.Sprintf("Error parsing settings logger file: %v", err))
	}

	var loggersKey string
	var formatter interface{}
	if appConfig.Environment == appConfig.Production {
		loggersKey = "init_loggers_prod"
		formatter = &logrus.JSONFormatter{}
	} else {
		loggersKey = "init_loggers_dev"
		formatter = &easy.Formatter{
			TimestampFormat: "2006-01-02 15:04:05",
			LogFormat:       "[%lvl%]: %time% - %msg%",
		}
	}

	err = os.MkdirAll(appConfig.AppRoot+"/logs", 0777)
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
				file, err := os.OpenFile(appConfig.AppRoot+fmt.Sprintf("/logs/%s", handler["filename"].(string)), os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
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
		loggers = append(loggers, &Logger{Logger: logger, name: key.(string)})
	}

	LogflnIfExists("info", "%d Loggers initialized", logrus.InfoLevel, LoggerLevelImportant, len(loggers))
}
