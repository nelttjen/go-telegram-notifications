package telegram

import (
	"context"
	libBot "github.com/go-telegram/bot"
	"notifications-bot/internal/services"
	"notifications-bot/pkg/driver/sql"
	"notifications-bot/pkg/logger"
	"time"
)

type Bot interface {
	SendMessage(chatID uint64, text string) error
	InitBot() error
	RunExecutor()
}

type bot struct {
	ctx          context.Context
	token        string
	BotChanel    *chan map[string]interface{}
	Api          *libBot.Bot
	DbConnection sql.Connection
}

func NewBot(ctx context.Context, dbConn sql.Connection, token string, botChannel *chan map[string]interface{}) (Bot, error) {
	opts := []libBot.Option{}
	goTgBot, err := libBot.New(token, opts...)

	if err != nil {
		return nil, err
	}

	return &bot{
		DbConnection: dbConn,
		ctx:          ctx,
		token:        token,
		BotChanel:    botChannel,
		Api:          goTgBot,
	}, nil
}

func (b *bot) InitBot() error {
	return nil
}

func (b *bot) RunExecutor() {
	db, _ := b.DbConnection.GetDBConnection()

	service, _ := services.NewNotificationService(db)

	for {
		select {
		case message := <-*b.BotChanel:
			logger.FastDebug("Received message from telegram bot: %v", message)
			go func() {
				err := b.SendMessage(message["chat_id"].(uint64), message["message"].(string))
				if err != nil {
					logger.FastWarn("Error sending message %v", err)

					service.CreateNewBannedChat(message["chat_id"].(uint64))

					logger.FastDebug("Added banned chat %v", message["chat_id"].(uint64))

				} else {
					b.markMessageAsSent(message)
				}
			}()
		default:
			//logger.FastDebug("No new messages")
			time.Sleep(time.Second * 1)
		}
	}

}
