package services

import (
	"gorm.io/gorm"
	"notifications-bot/internal/config"
	"notifications-bot/internal/models"
	repo "notifications-bot/internal/repositories"
	"notifications-bot/internal/rpc"
	"notifications-bot/pkg/logger"
	"time"
)

type NoNotificationsToCreateError struct{}

func (e NoNotificationsToCreateError) Error() string {
	return "No notifications to create"
}

type NotificationService struct {
	Repository            *repo.NotificationRepository
	TelegramBotRepository *repo.TelegramBotRepository
}

func NewNotificationService(db *gorm.DB) (*NotificationService, error) {
	repository, err := repo.NewNotificationRepository(db)

	if err != nil {
		return nil, err
	}

	telegramBotRepository, err := repo.NewTelegramBotRepository(db)
	if err != nil {
		return nil, err
	}

	return &NotificationService{
		Repository:            repository,
		TelegramBotRepository: telegramBotRepository,
	}, nil
}

func (s *NotificationService) GetUnprocessedNotifications(limit int) ([]*models.TelegramNotification, error) {
	return s.Repository.GetUnprocessedNotifications(limit)
}

func (s *NotificationService) SendNotification(botChannels map[string]*chan map[string]interface{}, notification *models.TelegramNotification) {
	if !notification.Bot.Enabled {
		return
	}

	botHost := notification.Bot.BotHost

	botChannel, ok := botChannels[botHost]
	if !ok {
		return
	}

	logger.FastDebug("Sending notification to bot: %s", botHost)
	*botChannel <- map[string]interface{}{
		"message":     notification.Text.Text,
		"chat_id":     notification.ChatID,
		"database_id": notification.ID,
	}

}

func (s *NotificationService) ProcessNotifications(botChannels map[string]*chan map[string]interface{}, items []*models.TelegramNotification) {
	if len(items) > 0 {
		logger.FastDebug("Found %d unprocessed notifications", len(items))
		for _, item := range items {
			go s.SendNotification(botChannels, item)
		}
		s.MarkNotificationsAsProcessed(items)
	}
}

func (s *NotificationService) MarkNotificationsAsProcessed(items []*models.TelegramNotification) {
	err := s.Repository.MarkNotificationsAsProcessed(items)
	if err != nil {
		logger.FastWarn("Failed to mark notifications as processed: %v", err)
	}
}

func (s *NotificationService) DatabaseCheckerProcess(botChannels map[string]*chan map[string]interface{}) {
	logger.FastInfo("Starting database checker process")
	for {
		time.Sleep(1 * time.Second)

		items, err := s.GetUnprocessedNotifications(config.NotificationPerSecondLimit)
		if err != nil {
			logger.FastWarn("Failed to get unprocessed notifications: %v", err)
			continue
		}

		s.ProcessNotifications(botChannels, items)
	}
}

func (s *NotificationService) AddNotificationsToQueue(request *rpc.AddNotificationsToQueueRequest) (uint64, error) {
	bannedChats, err := s.Repository.GetBannedChats()
	if err != nil {
		return 0, err
	}

	textID, err := s.Repository.CreateTextModel(request.Text)
	if err != nil {
		return 0, err
	}

	bots, err := s.TelegramBotRepository.GetTelegramBotsMapping()
	if err != nil {
		return 0, err
	}

	notifications := s.processRequestNotifications(request, bots, bannedChats, textID)

	if len(notifications) == 0 {
		return 0, NoNotificationsToCreateError{}
	}

	err = s.Repository.CreateNotifications(notifications, config.NotificationBatchSize)
	if err != nil {
		return 0, err
	}

	return textID, nil
}

func (s *NotificationService) CreateNewBannedChat(chatID uint64) {
	exists := s.TelegramBotRepository.GetBannedChat(chatID)
	if exists == nil {
		err := s.TelegramBotRepository.CreateNewBannedChat(chatID)
		if err != nil {
			logger.FastWarn("Failed to create new banned chat: %v", err)
		}
	}
}

func (s *NotificationService) GetRequestStatisticsResponse(requestID uint64) (*rpc.GetTelegramNotificationStatisticsResponse, error) {
	statistics, err := s.Repository.GetRequestStatistics(requestID)
	if err != nil {
		return nil, err
	}

	return &rpc.GetTelegramNotificationStatisticsResponse{
		Processed:         statistics.CountProcessed == statistics.CountTotal && statistics.CountTotal > 0,
		TotalMessages:     statistics.CountTotal,
		SentMessages:      statistics.CountSent,
		ProcessedMessages: statistics.CountProcessed,
		QueueMessages:     statistics.CountTotal - statistics.CountProcessed,
	}, nil
}

func (s *NotificationService) processRequestNotifications(
	request *rpc.AddNotificationsToQueueRequest,
	bots map[string]uint64,
	bannedChats map[uint64]any,
	textID uint64,
) []*models.TelegramNotification {
	var notifications []*models.TelegramNotification

	for _, item := range request.MessageSettings {
		_, ok := bannedChats[item.TelegramUserId]
		if ok {
			continue
		}
		botId, ok := bots[item.TelegramBotHost]
		if !ok {
			continue
		}
		notifications = append(notifications, &models.TelegramNotification{
			ChatID: item.TelegramUserId,
			BotID:  &botId,
			TextID: &textID,
		})
	}

	return notifications
}
