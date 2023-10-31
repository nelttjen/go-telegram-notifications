package app

import (
	"context"
	"remanga-notifications-bot/internal/rpc"
	"remanga-notifications-bot/pkg/logger"
)

type NotificationService struct {
	rpc.UnimplementedNotificationsServiceServer
}

func (s *NotificationService) AddNotificationsToQueue(ctx context.Context, request *rpc.AddNotificationsToQueueRequest) (*rpc.AddNotificationsToQueueResponse, error) {
	logger.FastDebug(request.String())
	return &rpc.AddNotificationsToQueueResponse{
		RequestId: 0,
		Success:   false,
	}, nil
}
