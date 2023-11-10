package telegram

import (
	"context"
	"github.com/sirupsen/logrus"
	"notifications-bot/internal/models"
	"notifications-bot/pkg/driver/sql"
	"notifications-bot/pkg/logger"
)

func InitializeBots(ctx context.Context) (map[string]*chan map[string]interface{}, error) {
	postgres := sql.NewPostgresConnection()
	err := postgres.MakeConnection()
	if err != nil {
		logger.LogflnIfExists("error", "Failed to connect to postgres: %v", logrus.FatalLevel, logger.LoggerLevelImportant, err)
		return nil, err
	}
	connection, _ := postgres.GetDBConnection()

	var bots []*models.TelegramBot

	connection.Model(&models.TelegramBot{}).Where(models.TelegramBot{Enabled: true}).Scan(&bots)

	botChannels := map[string]*chan map[string]interface{}{}

	for _, bot := range bots {
		botChan := make(chan map[string]interface{}, 1_000_000)
		logger.FastInfo("initializing bot %s", bot.BotHost)

		dbConn := sql.NewPostgresConnection()
		err := dbConn.MakeConnection()
		if err != nil {
			logger.FastWarn("Error getting database connection while creating bot %v", err)
			continue
		}

		botInstance, err := NewBot(ctx, dbConn, bot.BotToken, &botChan)
		if err != nil {
			logger.LogflnIfExists("info", "Failed to create new bot %s: %v", logrus.WarnLevel, logger.LoggerLevelImportant, bot.BotHost, err)
			continue
		}
		err = botInstance.InitBot()
		if err != nil {
			logger.LogflnIfExists("info", "Failed to init bot %s: %v", logrus.WarnLevel, logger.LoggerLevelImportant, bot.BotHost, err)
			continue
		}
		go botInstance.RunExecutor()
		botChannels[bot.BotHost] = &botChan
		logger.FastInfo("bot %s initialized", bot.BotHost)
	}

	return botChannels, nil
}
