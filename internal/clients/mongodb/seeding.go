package mongodb

import (
	"context"
	"fmt"
)

// Drop drops the entire database. This should be a noop in production, but can be useful for resetting the database in tests.
func (dbc *Database) Drop(ctx context.Context) error {

	if err := dbc.Collections.Notifications.Drop(ctx); err != nil {
		return fmt.Errorf("failed to drop notifications collection: %w", err)
	}

	if err := dbc.Collections.Alarms.Drop(ctx); err != nil {
		return fmt.Errorf("failed to drop alarms collection: %w", err)
	}

	if err := dbc.Collections.Maintenance.Drop(ctx); err != nil {
		return fmt.Errorf("failed to drop maintenance collection: %w", err)
	}

	if err := dbc.Collections.MaintenanceEditLogs.Drop(ctx); err != nil {
		return fmt.Errorf("failed to drop maintenance logs collection: %w", err)
	}

	if err := dbc.Collections.ACLs.Drop(ctx); err != nil {
		return fmt.Errorf("failed to drop ACLs collection: %w", err)
	}

	return nil
}
