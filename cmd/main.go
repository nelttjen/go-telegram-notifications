package main

import (
	"github.com/sirupsen/logrus"
	"remanga-notifications-bot/internal/app"
	"remanga-notifications-bot/internal/config"
	"remanga-notifications-bot/internal/models"
	"remanga-notifications-bot/internal/telegram"
	"remanga-notifications-bot/pkg/driver/sql"
	"remanga-notifications-bot/pkg/logger"
)

func main() {
	logger.InitLoggers()
	logger.FastDebug("Starting app")

	newApp := app.NewApp(config.EnvRoot)
	err := newApp.Init()

	if err != nil {
		logger.LogflnIfExists("error", "Failed to init app: %v", logrus.FatalLevel, logger.LoggerLevelImportant, err)
		panic(err)
	}

	postgres := sql.NewPostgresConnection()
	err = postgres.MakeConnection()
	if err != nil {
		logger.LogflnIfExists("error", "Failed to connect to postgres: %v", logrus.FatalLevel, logger.LoggerLevelImportant, err)
		panic(err)
	}
	connection, _ := postgres.GetDBConnection()

	var bots []*models.TelegramBot

	connection.Model(&models.TelegramBot{}).Where(models.TelegramBot{Enabled: true}).Scan(&bots)

	for _, bot := range bots {
		logger.FastInfo("initializing bot %s", bot.BotHost)
		botInstance := telegram.NewBot(bot.BotToken)
		err := botInstance.InitBot()
		if err != nil {
			logger.LogflnIfExists("info", "Failed to init bot %s: %v", logrus.WarnLevel, logger.LoggerLevelImportant, bot.BotHost, err)
		}
		go botInstance.RunExecutor()
		logger.FastInfo("bot %s initialized", bot.BotHost)
	}

	if err := newApp.Run(); err != nil {
		logger.LogflnIfExists("error", "Failed to run app: %v", logrus.FatalLevel, logger.LoggerLevelImportant, err)
		panic(err)
	}
	logger.LoglnIfExists("info", "App stopped", logrus.InfoLevel, logger.LoggerLevelImportant)
}
