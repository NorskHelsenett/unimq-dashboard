package mongodb

import (
	"github.com/sisneve/rabbitmq-dashboard/internal/database"
	"github.com/sisneve/rabbitmq-dashboard/internal/models"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type MaintenaceLogHandler struct {
	Collection mongo.Collection
	CollectorHolder[models.MaintenanceEditLog]
}

func NewMaintenanceLogHandler(collection mongo.Collection) *MaintenaceLogHandler {
	return &MaintenaceLogHandler{
		Collection:      collection,
		CollectorHolder: CollectorHolder[models.MaintenanceEditLog]{},
	}
}

// interface guard clause.
var _ database.MaintenaceLogRepository = (*MaintenaceLogHandler)(nil)
