package env

import (
	"fmt"
	"github.com/joho/godotenv"
	"os"
	"strconv"
)

var _ Env = &env{}

type env struct {
	EnvFilePath string
}

type KeyEnvError struct {
	key string
}

type TypeEnvError struct {
	value        interface{}
	expectedType string
}

func (e *KeyEnvError) Error() string {
	return fmt.Sprintf("No such key in environment: %s, default value was not provided", e.key)
}

func (e *TypeEnvError) Error() string {
	return fmt.Sprintf("Failed to convert type to expected type %s, got value: %s", e.expectedType, e.value)
}

type Env interface {
	LoadEnv() error
	GetEnv(key string, defaultValue ...interface{}) (interface{}, error)
	GetEnvAsString(key string, defaultValue ...string) (string, error)
	GetEnvAsInt(key string, defaultValue ...int) (int, error)
	GetEnvAsBool(key string, defaultValue ...bool) (bool, error)
}

func (e *env) LoadEnv() error {
	_, err := os.Stat(e.EnvFilePath)

	if os.IsNotExist(err) {
		return err
	}

	if err := godotenv.Load(e.EnvFilePath); err != nil {
		return err
	}

	return nil
}

func (e *env) GetEnv(key string, defaultValue ...interface{}) (interface{}, error) {
	val, exists := os.LookupEnv(key)

	if !exists {
		if defaultValue != nil {
			return defaultValue[0], nil
		}
		return "", &KeyEnvError{key: key}
	}
	return val, nil
}

func (e *env) GetEnvAsString(key string, defaultValue ...string) (string, error) {
	val, err := e.GetEnv(key)

	if err != nil {
		if defaultValue != nil {
			return defaultValue[0], nil
		}
		return "", &KeyEnvError{key: key}
	}
	return val.(string), nil
}

func (e *env) GetEnvAsInt(key string, defaultValue ...int) (int, error) {
	val, err := e.GetEnv(key)
	if err != nil {
		if defaultValue != nil {
			return defaultValue[0], nil
		}
		return 0, err
	}
	intVal, err := strconv.Atoi(val.(string))
	if err != nil {
		if defaultValue != nil {
			return defaultValue[0], nil
		}
		return 0, &TypeEnvError{value: val, expectedType: "int"}
	}
	return intVal, nil
}

func (e *env) GetEnvAsBool(key string, defaultValue ...bool) (bool, error) {
	val, err := e.GetEnv(key)
	if err != nil {
		if defaultValue != nil {
			return defaultValue[0], nil
		}
		return false, err
	}
	boolVal, err := strconv.ParseBool(val.(string))
	if err != nil {
		if defaultValue != nil {
			return defaultValue[0], nil
		}
		return false, &TypeEnvError{value: val, expectedType: "bool"}
	}
	return boolVal, nil
}

func NewEnv(filepath string) Env {
	newEnv := env{
		EnvFilePath: filepath,
	}
	return &newEnv
}
