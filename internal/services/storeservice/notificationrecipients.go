package storeservice

import (
	"context"

	"github.com/sisneve/rabbitmq-dashboard/internal/database"
	"github.com/sisneve/rabbitmq-dashboard/internal/models"
)

func (ss *StoreService) GetNotificationRecipientsByParentID(ctx context.Context, parentID string) ([]*models.Recipient, error) {
	filter := database.NewFilter("_id", database.FilterOpEqual, parentID)
	recipient, err := ss.Store.NotificationRecipientsHandler().Read(ctx, filter)
	if err != nil {
		return nil, err
	}

	return recipient, nil
}

func (ss *StoreService) GetNotificationRecipientsByID(ctx context.Context, parentID, id string) (*models.Recipient, error) {
	recipients, err := ss.GetNotificationRecipientsByParentID(ctx, parentID)
	if err != nil {
		return nil, err
	}

	for _, recipient := range recipients {
		if recipient.ID == id {
			return recipient, nil
		}
	}

	return nil, database.ErrResourceNotFound
}
