package storeservice

import (
	"context"

	"github.com/sisneve/rabbitmq-dashboard/internal/models"
)

func (rc *StoreService) SeedACLs(ctx context.Context, acls []*models.ACL) error {
	if err := rc.ACLHandler().Put(ctx, acls); err != nil {
		return err
	}
	return nil
}
