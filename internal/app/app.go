package app

import (
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"net"
	"remanga-notifications-bot/pkg/driver/sql"
	"remanga-notifications-bot/pkg/env"
	"remanga-notifications-bot/pkg/logger"

	"remanga-notifications-bot/internal/config"
	"remanga-notifications-bot/internal/models"
	"remanga-notifications-bot/internal/rpc"
)

var _ App = &app{}

type app struct {
	server *grpc.Server
	env    env.Env

	initialized bool
}

type App interface {
	Init() error
	Run() error
}

func NewApp(envPath string) App {
	server := grpc.NewServer()
	newEnv := env.NewEnv(envPath)

	newApp := app{
		server:      server,
		env:         newEnv,
		initialized: false,
	}
	return &newApp
}

func (a *app) Init() error {
	err := a.env.LoadEnv()
	if err != nil {
		logger.LogflnIfExists("error", "Failed to load .env file: %v", logrus.FatalLevel, logger.LoggerLevelImportant, err)
		return err
	}
	postgres := sql.NewPostgresConnection()
	err = postgres.MakeConnection()
	if err != nil {
		return err
	}

	connection, _ := postgres.GetDBConnection()
	err = connection.AutoMigrate(&models.TelegramMessageNotification{}, &models.TelegramNotification{}, &models.TelegramBot{})

	if err != nil {
		logger.LogflnIfExists("error", "Failed to migrate tables: %v", logrus.FatalLevel, logger.LoggerLevelImportant, err)
		return err
	}

	rpc.RegisterNotificationsServiceServer(a.server, &NotificationService{})
	logger.LoglnIfExists("info", "App initialization done", logrus.InfoLevel, logger.LoggerLevelImportant)

	a.initialized = true

	return nil
}

func (a *app) Run() (err error) {
	if !a.initialized {
		panic("App is not initialized. Call Init() before Run()")
	}

	lis, err := net.Listen(config.Protocol, config.Port)
	if err != nil {
		return err
	}

	if err := a.server.Serve(lis); err != nil {
		return err
	}

	return nil
}
