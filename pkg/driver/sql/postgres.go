package sql

import (
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"notifications-bot/internal/config"
	"notifications-bot/pkg/env"
	"notifications-bot/pkg/logger"
)

var _ Connection = &postgresConnection{}

type postgresConnection struct {
	host     string
	port     int
	user     string
	password string
	database string

	DBConnection  *gorm.DB
	DBTransaction *gorm.DB
}

func NewPostgresConnection() Connection {
	newEnv := env.NewEnv(config.EnvRoot)

	host, _ := newEnv.GetEnvAsString("POSTGRES_HOST", "127.0.0.1")
	port, _ := newEnv.GetEnvAsInt("POSTGRES_PORT", 5432)
	user, _ := newEnv.GetEnvAsString("POSTGRES_USER", "postgres")
	password, _ := newEnv.GetEnvAsString("POSTGRES_PASSWORD", "postgres")
	database, _ := newEnv.GetEnvAsString("POSTGRES_DB", "default")

	conn := postgresConnection{
		host:     host,
		port:     port,
		user:     user,
		password: password,
		database: database,
	}
	return &conn
}

func (c *postgresConnection) formatConnection(redactPassword bool) string {
	password := "<HIDDEN>"

	if !redactPassword {
		password = c.password
	}

	return fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=Europe/Moscow",
		c.host, c.user, password, c.database, c.port)
}

func (c *postgresConnection) MakeConnection() error {
	logger.FastTrace("Connecting to database with connection string: %s", c.formatConnection(true))

	conn, err := gorm.Open(postgres.Open(c.formatConnection(false)), &gorm.Config{})
	if err != nil {
		return err
	}
	c.DBConnection = conn

	logger.FastTrace("Connected to postgres database")

	return nil
}

func (c *postgresConnection) NewTransaction() error {
	if c.DBConnection == nil {
		return &NoConnectionError{}
	}

	if c.DBTransaction != nil {
		return &TransactionExists{}
	}

	c.DBTransaction = c.DBConnection.Begin()

	return nil
}

func (c *postgresConnection) CommitTransaction(endTransaction bool) error {
	if c.DBTransaction == nil {
		return &NoTransactionError{}
	}

	c.DBTransaction.Commit()

	if endTransaction {
		c.DBTransaction = nil
	}
	return nil
}

func (c *postgresConnection) RollbackTransaction(endTransaction bool) error {
	if c.DBTransaction == nil {
		return &NoTransactionError{}
	}

	c.DBTransaction.Rollback()

	if endTransaction {
		c.DBTransaction = nil
	}
	return nil
}

func (c *postgresConnection) ExecTransaction(operation string, values []interface{}) error {
	if c.DBTransaction == nil {
		return &NoTransactionError{}
	}
	c.DBTransaction.Exec(operation, values...)
	if c.DBTransaction.Error != nil {
		return c.DBTransaction.Error
	}
	return nil
}

func (c *postgresConnection) GetDBConnection() (*gorm.DB, error) {

	if c.DBConnection == nil {
		return nil, &NoConnectionError{}
	}
	return c.DBConnection, nil
}

func (c *postgresConnection) GetDBTransaction() (*gorm.DB, error) {
	if c.DBTransaction == nil {
		return nil, &NoTransactionError{}
	}
	return c.DBTransaction, nil
}
