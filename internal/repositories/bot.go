package repositories

import (
	"gorm.io/gorm"
	"notifications-bot/internal/config"
	"notifications-bot/internal/models"
	"notifications-bot/pkg/driver/sql"
	"time"
)

type TelegramBotRepository struct {
	Database *gorm.DB
}

func NewTelegramBotRepository(db *gorm.DB) (*TelegramBotRepository, error) {
	if db == nil {
		connection := sql.NewPostgresConnection()
		err := connection.MakeConnection()
		if err != nil {
			return nil, err
		}

		db, _ = connection.GetDBConnection()
	}

	return &TelegramBotRepository{
		Database: db,
	}, nil
}

func (repo *TelegramBotRepository) GetTelegramBotsMapping() (map[string]uint64, error) {
	var bots []*models.TelegramBot

	result := repo.Database.Model(&models.TelegramBot{}).Where("enabled =?", true).Find(&bots)

	botMapping := make(map[string]uint64)

	if result.Error == nil {
		for _, item := range bots {
			botMapping[item.BotHost] = item.ID
		}
	}

	return botMapping, result.Error
}

func (repo *TelegramBotRepository) GetBannedChat(chatID uint64) *models.BannedChats {
	var bannedChat *models.BannedChats

	repo.Database.Model(&models.BannedChats{}).Where(
		"chat_id = ? AND until > ?",
		chatID, time.Now(),
	).Scan(&bannedChat)

	return bannedChat
}

func (repo *TelegramBotRepository) CreateNewBannedChat(chatID uint64) error {
	until := time.Now().Add(time.Duration(24) * time.Duration(config.BannedChatDays) * time.Hour)
	result := repo.Database.Model(&models.BannedChats{}).Create(&models.BannedChats{
		ChatID: chatID,
		Until:  until,
	})
	return result.Error
}
