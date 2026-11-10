package mongodb

import (
	"github.com/sisneve/rabbitmq-dashboard/internal/database"
	"github.com/sisneve/rabbitmq-dashboard/internal/models"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type NotificationHandler struct {
	Collection *mongo.Collection
	*CollectorHolder[models.VhostNotification]
}

func NewNotificationHandler(collection *mongo.Collection) *NotificationHandler {
	return &NotificationHandler{
		Collection:      collection,
		CollectorHolder: NewCollectorHolder[models.VhostNotification](collection),
	}
}

// interface guard clause.
var _ database.NotificationRepository = (*NotificationHandler)(nil)
