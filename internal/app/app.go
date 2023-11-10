package app

import (
	"context"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"net"
	"notifications-bot/internal/services"
	"notifications-bot/pkg/driver/sql"
	"notifications-bot/pkg/logger"

	"notifications-bot/internal/config"
	"notifications-bot/internal/models"
	"notifications-bot/internal/rpc"
)

var _ App = &app{}

type app struct {
	server      *grpc.Server
	BotChannels map[string]*chan map[string]interface{}
	BotContext  context.Context
	BotCancel   context.CancelFunc

	initialized bool
}

type App interface {
	Init() error
	Run() error
}

func NewApp(botChannels map[string]*chan map[string]interface{}, botContext context.Context, botCancel context.CancelFunc) App {
	server := grpc.NewServer()

	newApp := app{
		server:      server,
		BotChannels: botChannels,
		BotContext:  botContext,
		BotCancel:   botCancel,
		initialized: false,
	}
	return &newApp
}

func (a *app) Init() error {
	err := a.makeMigrations()
	if err != nil {
		logger.LogflnIfExists("error", "Failed to migrate tables: %v", logrus.FatalLevel, logger.LoggerLevelImportant, err)
		return err
	}

	rpc.RegisterNotificationsServiceServer(a.server, &NotificationService{})

	err = a.runCheckerProcess()
	if err != nil {
		logger.LogflnIfExists("error", "Failed to run checker process: %v", logrus.FatalLevel, logger.LoggerLevelImportant, err)
		return err
	}

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

func (a *app) makeMigrations() error {
	postgres := sql.NewPostgresConnection()
	err := postgres.MakeConnection()
	if err != nil {
		return err
	}

	connection, _ := postgres.GetDBConnection()
	err = connection.AutoMigrate(
		&models.TelegramMessageNotification{}, &models.TelegramNotification{},
		&models.TelegramBot{}, &models.BannedChats{},
	)
	return err
}

func (a *app) runCheckerProcess() error {
	notificationService, err := services.NewNotificationService(nil)
	if err != nil {
		logger.LogflnIfExists("error", "Failed to create notification service: %v", logrus.FatalLevel, logger.LoggerLevelImportant, err)
		return err
	}
	go notificationService.DatabaseCheckerProcess(a.BotChannels)
	return nil
}
