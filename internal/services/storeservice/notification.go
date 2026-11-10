package storeservice

import (
	"context"
	"fmt"

	"github.com/sisneve/rabbitmq-dashboard/internal/database"
	"github.com/sisneve/rabbitmq-dashboard/internal/models"
)

func (ss *StoreService) SeedNotifications(ctx context.Context, entries []*models.VhostNotification) error {
	return fmt.Errorf("SeedNotifications not implemented yet")
}

func (ss *StoreService) GetNotificationsByID(ctx context.Context, id string) (*models.VhostNotification, error) {
	filter := database.NewFilter("_id", database.FilterOpEqual, id)
	notifications, err := ss.Store.NotificationHandler().Read(ctx, filter)
	if err != nil {
		return nil, err
	}

	if len(notifications) == 0 {
		return nil, fmt.Errorf("no notifications found for id: %s", id)
	}

	if len(notifications) > 1 {
		return nil, fmt.Errorf("multiple notifications found for id: %s", id)
	}

	return notifications[0], nil
}

func (ss *StoreService) DeleteNotificationByID(ctx context.Context, id string) error {
	filter := database.NewFilter("_id", database.FilterOpEqual, id)
	err := ss.Store.NotificationHandler().Delete(ctx, filter)
	if err != nil {
		return err
	}

	return nil
}
