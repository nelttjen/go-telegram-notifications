package config

// log levels
const (
	LoggerLevelNone = iota
	LoggerLevelImportant
	LoggerLevelDebug
	LoggerLevelTrace
)

const (
	LogDisabled = iota
	LogEnabled
)

const (
	DebugLoggerName = "debug"
	InfoLoggerName  = "info"
	ErrorLoggerName = "error"
)
