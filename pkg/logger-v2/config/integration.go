package config

type LoggingConfig interface {
	GetLoggerEnabled() int
	GetLoggerLevel() int
	IsDevEnvironment() bool
	GetAppRoot() string
}
