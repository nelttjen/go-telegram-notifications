package sql

import (
	"fmt"
	"gorm.io/gorm"
)

type NoConnectionError struct{}

type NoTransactionError struct{}

type TransactionExists struct{}

func (e *NoConnectionError) Error() string {
	return fmt.Sprintf("There's no active connection.")
}

func (e *NoTransactionError) Error() string {
	return fmt.Sprintf("There's no active transaction.")
}

func (e *TransactionExists) Error() string {
	return fmt.Sprintf("There's already active transaction.")
}

type Connection interface {
	MakeConnection() error
	NewTransaction() error
	ExecTransaction(operation string, values []interface{}) error
	CommitTransaction(endTransaction bool) error
	RollbackTransaction(endTransaction bool) error

	GetDBConnection() (*gorm.DB, error)
	GetDBTransaction() (*gorm.DB, error)
}
