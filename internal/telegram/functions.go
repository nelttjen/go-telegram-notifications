package telegram

import (
	libBot "github.com/go-telegram/bot"
	"notifications-bot/internal/models"
)

func (b *bot) SendMessage(chatID uint64, text string) error {
	_, err := b.Api.SendMessage(b.ctx, &libBot.SendMessageParams{
		ChatID: chatID,
		Text:   text,
	})
	return err
}

func (b *bot) markMessageAsSent(message map[string]interface{}) {
	dbId := message["database_id"]

	conn, _ := b.DbConnection.GetDBConnection()
	conn.Model(&models.TelegramNotification{}).Where("id = ?", dbId).Update("sent", true)
}
