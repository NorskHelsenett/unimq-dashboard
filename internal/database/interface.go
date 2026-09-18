package database

import (
	"context"

	"github.com/sisneve/rabbitmq-dashboard/internal/models"
)

// Playground for new database interface. This comment will be removed once interface is fully implemented and a structure has been decided.

// AlarmHandler defines the interface for managing alarms in the database.
//
// Alarms are parented to an ID and can have multiple log entries.
type AlarmHandler interface {
	Store[models.AlarmEntry]
	Append(ctx context.Context, alarmID string, logEntry models.LogEntry) error
}

// TODO: Someone else is handling the ACL implementation, this is just a placeholder for now.
type ACLHandler interface {
}

// VhostHandler defines the interface for managing vhosts in the database.
//
// Vhosts are simply a base of CRUD operations for the vhost model.
type VhostHandler interface {
	Store[models.Vhost]
}

// MaintenanceHandler defines the interface for managing maintenance entries in the database.
//
// Maintenance entries are parented to an ID and can have multiple log entries.
type MaintenaceHandler interface {
	Advance(ctx context.Context) error
	Store[models.MaintenanceEntry]
}

type NotificationHandler interface {
	Store[models.VhostNotification]
	Toggle(ctx context.Context, id string) error
	GetByParentID(ctx context.Context, parentID string) ([]*models.VhostNotification, error)
}

// This is a base interface for most database operations. It defines the basic CRUD operations for a generic type T.
type Store[T any] interface {
	GetAll(ctx context.Context) ([]*T, error)
	GetByID(ctx context.Context, id string) (*T, error)
	Create(ctx context.Context, item *T) error
	Update(ctx context.Context, id string, item *T) error
	Delete(ctx context.Context, id string) error
}
