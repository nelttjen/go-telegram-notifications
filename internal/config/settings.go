package config

import (
	"path/filepath"
	"runtime"
)

var (
	_, dir, _, _ = runtime.Caller(0)
	AppRoot      = filepath.Dir(filepath.Dir(filepath.Dir(dir)))
	EnvRoot      = AppRoot + "/.env"

	Production  = "prod"
	Development = "dev"
	Testing     = "test"

	Protocol = "tcp"
	Port     = ":55000"
)

var Environment = Development
