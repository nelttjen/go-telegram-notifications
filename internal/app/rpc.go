package app

import (
	"context"
	"errors"
	"notifications-bot/internal/rpc"
	"notifications-bot/internal/services"
	"notifications-bot/pkg/logger"
)

type NotificationService struct {
	rpc.UnimplementedNotificationsServiceServer
}

func (s *NotificationService) AddNotificationsToQueue(ctx context.Context, request *rpc.AddNotificationsToQueueRequest) (*rpc.AddNotificationsToQueueResponse, error) {
	logger.FastDebug(request.String())

	service, err := services.NewNotificationService(nil)
	if err != nil {
		return &rpc.AddNotificationsToQueueResponse{
			RequestId: 0,
			Success:   false,
			Message:   "internal server error",
		}, err
	}

	requestID, err := service.AddNotificationsToQueue(request)

	if err != nil {
		if errors.As(err, &services.NoNotificationsToCreateError{}) {
			return &rpc.AddNotificationsToQueueResponse{
				RequestId: 0,
				Success:   false,
				Message:   err.Error(),
			}, nil
		} else {
			println(errors.As(err, &services.NoNotificationsToCreateError{}))
			return &rpc.AddNotificationsToQueueResponse{
				RequestId: 0,
				Success:   false,
				Message:   "internal server error",
			}, nil
		}
	}

	return &rpc.AddNotificationsToQueueResponse{
		RequestId: requestID,
		Success:   true,
		Message:   "ok",
	}, nil
}

func (s *NotificationService) GetTelegramNotificationStatistics(ctx context.Context, request *rpc.GetTelegramNotificationStatisticsRequest) (*rpc.GetTelegramNotificationStatisticsResponse, error) {
	service, err := services.NewNotificationService(nil)

	if err != nil {
		return &rpc.GetTelegramNotificationStatisticsResponse{}, nil
	}
	response, err := service.GetRequestStatisticsResponse(request.RequestId)

	if err != nil {
		logger.FastError("Error getting request statistics: %v", err)
	}

	return response, nil
}
