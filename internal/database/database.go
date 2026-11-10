package database

import (
	"context"
	"errors"

	"github.com/sisneve/rabbitmq-dashboard/internal/models"
)

type Store interface {
	InitCollections() error
	Ping(ctx context.Context, timeoutSecs int) error

	AlarmHandler() AlarmRepository
	ACLHandler() ACLRepository
	MaintenanceHandler() MaintenaceRepository
	MaintenanceLogsHandler() MaintenaceLogRepository
	// VhostHandler() VhostHandler
	NotificationHandler() NotificationRepository
	NotificationRecipientsHandler() NotificationRecipientRepository
	NotificationRulesHandler() NotificationRulesRepository
}

type StoreSeeder interface {
	Seed(ctx context.Context) error
}

// AlarmRepository defines the interface for managing alarms in the database.
type AlarmRepository interface {
	CRUDStore[models.AlarmEntry]
	Append(ctx context.Context, alarmID string, logEntry models.LogEntry) error
}

// ACLRepository defines the interface for managing ACLs in the database.
type ACLRepository interface {
	CRUDStore[models.ACL]
	Putter[models.ACL]
	Patcher[models.ACL]
	GetByGroupIDs(ctx context.Context, groupIDs []string) ([]*models.ACL, error)
}

// VhostHandler defines the interface for managing vhosts in the database.
type VhostRepository interface {
	CRUDStore[models.Vhost]
}

// MaintenenaceRepository defines the interface for managing maintenance entries in the database.
type MaintenaceRepository interface {
	Advance(ctx context.Context) error
	CRUDStore[models.MaintenanceEntry]
	Patcher[models.PatchMaintenanceEntry]
}

// MaintenaceLogHandler defines the interface for managing maintenance edit logs in the database.
type MaintenaceLogRepository interface {
	CRUDStore[models.MaintenanceEditLog]
}

// NotificationRepository defines the interface for managing notifications in the database.
type NotificationRepository interface {
	CRUDStore[models.VhostNotification]
}

type NotificationRecipientRepository interface {
	CRUDStore[models.Recipient]
}

type NotificationRulesRepository interface {
	CRUDStore[models.AlarmRule]
}

// This is a base interface for most database operations. It defines the basic CRUD operations for a generic type T.
type CRUDStore[T any] interface {
	Create(ctx context.Context, items []*T) error                 // Create a single item.
	Read(ctx context.Context, filters ...Filter) ([]*T, error)    // Read one or more items based on filters.
	Update(ctx context.Context, item *T, filters ...Filter) error // Update a single item.
	Delete(ctx context.Context, filters ...Filter) error          // Delete one or more items based on filters.
	Drop(ctx context.Context) error                               // Drop the entire collection. This should be a noop operation in production, but useful for testing or resetting the database.
}

type Putter[T any] interface {
	Put(ctx context.Context, items []*T, filters ...Filter) error // Upsert an item (create or update).
}

type Patcher[T any] interface {
	Patch(ctx context.Context, patch map[string]any, filters ...Filter) error // Patch an item by ID.
}

type (
	Filter struct {
		Property string
		Type     FilterType
		Value    any
	}

	FilterType string
)

const (
	FilterOpEqual        FilterType = "eq"
	FilterOpNotEqual     FilterType = "ne"
	FilterOpIn           FilterType = "in"
	FilterOpNotIn        FilterType = "nin"
	FilterOpGreater      FilterType = "gt"
	FilterOpLess         FilterType = "lt"
	FilterOpGreaterEqual FilterType = "gte"
	FilterOpLessEqual    FilterType = "lte"
)

func NewFilter(property string, filterType FilterType, value any) Filter {
	return Filter{
		Property: property,
		Type:     filterType,
		Value:    value,
	}
}

type DatabaseMonitor interface {
	Ping(ctx context.Context, timeoutSecs int) error
}

var (
	ErrGeneralDatabaseError     = errors.New("general database error")
	ErrResourceNotFound         = errors.New("resource not found")
	ErrInvalidRequestParameters = errors.New("invalid request parameters")
)
