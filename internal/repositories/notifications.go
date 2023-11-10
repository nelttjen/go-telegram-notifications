package repositories

import (
	"gorm.io/gorm"
	"notifications-bot/internal/models"
	"notifications-bot/pkg/driver/sql"
	"time"
)

type NotificationRepository struct {
	Database *gorm.DB
}

type RequestStatistics struct {
	CountSent      uint32 `gorm:"column:countsent" json:"countsent"`
	CountProcessed uint32 `gorm:"column:countprocessed" json:"countprocessed"`
	CountTotal     uint32 `gorm:"column:counttotal" json:"counttotal"`
}

func NewNotificationRepository(db *gorm.DB) (*NotificationRepository, error) {
	if db == nil {
		connection := sql.NewPostgresConnection()
		err := connection.MakeConnection()
		if err != nil {
			return nil, err
		}

		db, _ = connection.GetDBConnection()
	}

	return &NotificationRepository{
		Database: db,
	}, nil
}

func (repo *NotificationRepository) GetUnprocessedNotifications(limit int) ([]*models.TelegramNotification, error) {
	var items []*models.TelegramNotification
	result := repo.Database.Model(&models.TelegramNotification{}).Joins("Text").Joins("Bot").Where("processed = ?", false).Limit(limit).Find(&items)
	if result.Error != nil {
		return nil, result.Error
	}
	return items, nil
}

func (repo *NotificationRepository) MarkNotificationsAsProcessed(items []*models.TelegramNotification) error {
	ids := make([]uint64, len(items))

	for _, item := range items {
		ids = append(ids, item.ID)
	}

	result := repo.Database.Model(&models.TelegramNotification{}).Where("id in (?)", ids).Update("processed", true)
	return result.Error
}

func (repo *NotificationRepository) GetBannedChats() (map[uint64]any, error) {
	var items []*models.BannedChats

	result := repo.Database.Model(&models.BannedChats{}).Where("until > ?", time.Now()).Find(&items)
	if result.Error != nil {
		return make(map[uint64]any), result.Error
	}

	var bannedChats = make(map[uint64]any)

	for _, item := range items {
		bannedChats[item.ChatID] = nil
	}

	return bannedChats, nil
}

func (repo *NotificationRepository) CreateTextModel(text string) (uint64, error) {
	newModel := &models.TelegramMessageNotification{
		Text: text,
	}
	result := repo.Database.Model(&models.TelegramMessageNotification{}).Create(newModel)
	return newModel.ID, result.Error
}

func (repo *NotificationRepository) CreateNotifications(notifications []*models.TelegramNotification, batchSize int) error {
	result := repo.Database.CreateInBatches(notifications, batchSize)
	return result.Error
}

func (repo *NotificationRepository) GetRequestStatistics(requestID uint64) (*RequestStatistics, error) {
	rawSql := `
		with countSent as (
			select count(*) as CountSent from telegram_notifications
			where sent = true and text_id = ?
		),
		countTotal as (
			select count(*) as CountTotal from telegram_notifications
			where text_id = ?
		),
		countProcessed as (
			select count(*) as CountProcessed from telegram_notifications
			where processed = true and text_id = ?
		)
		select 
		s.CountSent as CountSent,
        c.CountTotal as CountTotal,
        p.CountProcessed as CountProcessed
		from countSent s
		join countTotal c on true
		join countProcessed p on true;
	`
	result := &RequestStatistics{}
	dbResult := repo.Database.Raw(rawSql, requestID, requestID, requestID).Scan(result)
	return result, dbResult.Error
}
