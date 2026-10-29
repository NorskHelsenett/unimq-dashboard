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
	// VhostHandler() VhostHandler
	// MaintenanceHandler() MaintenaceHandler
	// MaintenanceLogsHandler() MaintenaceLogsHandler
	// NotificationHandler() NotificationHandler
	// NotificationRecipientHandler() NotificationRecipientHandler
	// NotificationRulesHandler() NotificationRulesHandler
}

type StoreSeeder interface {
	Seed(ctx context.Context) error
}

// AlarmHandler defines the interface for managing alarms in the database.
//
// Alarms are parented to an ID and can have multiple log entries.
type AlarmRepository interface {
	CRUDStore[models.AlarmEntry]
	Append(ctx context.Context, alarmID string, logEntry models.LogEntry) error
}

// ACLHandler defines the interface for managing ACLs in the database.
type ACLRepository interface {
	CRUDStore[models.ACL]
	CRUDStoreExtended[models.ACL]
	GetByGroupIDs(ctx context.Context, groupIDs []string) ([]*models.ACL, error)
}

// VhostHandler defines the interface for managing vhosts in the database.
type VhostHandler interface {
	CRUDStore[models.Vhost]
}

// MaintenanceHandler defines the interface for managing maintenance entries in the database.
type MaintenaceHandler interface {
	Advance(ctx context.Context) error
	CRUDStore[models.MaintenanceEntry]
}

type MaintenaceLogsHandler interface {
	CRUDStore[models.MaintenanceEditLog]
}

type NotificationHandler interface {
	CRUDStore[models.VhostNotification]
	Toggle(ctx context.Context, id string) error
	GetByParentID(ctx context.Context, parentID string) ([]*models.VhostNotification, error)
}

type NotificationRecipientHandler interface {
	CRUDStore[models.Recipient]
}

type NotificationRulesHandler interface {
	CRUDStore[models.AlarmRule]
}

// This is a base interface for most database operations. It defines the basic CRUD operations for a generic type T.
type CRUDStore[T any] interface {
	Create(ctx context.Context, item *T) error                 // Create a single item.
	Read(ctx context.Context, filters ...Filter) ([]*T, error) // Read one or more items based on filters.
	Update(ctx context.Context, item *T) error                 // Update a single item.
	Delete(ctx context.Context, filters ...Filter) error       // Delete one or more items based on filters.
}

// CRUDStoreExtended extends the CRUDStore interface with additional methods for upserting and patching items.
type CRUDStoreExtended[T any] interface {
	Put(ctx context.Context, item *T) error                             // Upsert an item (create or update).
	Patch(ctx context.Context, id string, updates map[string]any) error // Patch an item by ID.
}

type (
	Filter struct {
		Property string
		Type     FilterType
		Value    string
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

func NewFilter(property string, filterType FilterType, value string) Filter {
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
	ErrResourceNotFound = errors.New("resource not found")
)
