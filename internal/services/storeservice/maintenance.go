package storeservice

import (
	"context"

	"github.com/sisneve/rabbitmq-dashboard/internal/models"
)

func (ss *StoreService) SeedMaintenace(ctx context.Context, entries []*models.MaintenanceEntry) error {

	err := ss.Store.MaintenanceHandler().Create(ctx, entries)
	if err != nil {
		return err
	}
	return nil
}

func (ss *StoreService) SeedMaintenanceLogs(ctx context.Context, entries []*models.MaintenanceEditLog) error {

	err := ss.Store.MaintenanceLogsHandler().Create(ctx, entries)
	if err != nil {
		return err
	}

	return nil
}
