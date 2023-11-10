package main

import (
	"context"
	"github.com/sirupsen/logrus"
	"notifications-bot/internal/app"
	"notifications-bot/internal/config"
	"notifications-bot/internal/telegram"
	envLib "notifications-bot/pkg/env"
	"notifications-bot/pkg/logger"
)

func main() {
	logger.InitLoggers()
	logger.FastDebug("Starting app")

	env := envLib.NewEnv(config.EnvRoot)
	err := env.LoadEnv()

	if err != nil {
		logger.LogflnIfExists("error", "Failed to load .env file: %v", logrus.FatalLevel, logger.LoggerLevelImportant, err)
		panic(err)
	}

	botContext, cancel := context.WithCancel(context.TODO())

	botChannels, err := telegram.InitializeBots(botContext)

	if err != nil {
		logger.FastFatal("Cannot initialize bots %v", err)
		panic(err)
	}

	newApp := app.NewApp(botChannels, botContext, cancel)
	err = newApp.Init()

	if err != nil {
		logger.LogflnIfExists("error", "Failed to init app: %v", logrus.FatalLevel, logger.LoggerLevelImportant, err)
		panic(err)
	}

	if err := newApp.Run(); err != nil {
		logger.LogflnIfExists("error", "Failed to run app: %v", logrus.FatalLevel, logger.LoggerLevelImportant, err)
		panic(err)
	}
	logger.LoglnIfExists("info", "App stopped", logrus.InfoLevel, logger.LoggerLevelImportant)
}
