package storeservice

import (
	"context"

	"github.com/sisneve/rabbitmq-dashboard/internal/models"
)

func (rc *StoreService) SeedACLs(ctx context.Context, acls []models.ACL) error {
	for _, acl := range acls {
		if err := rc.Store.ACLHandler().Put(ctx, &acl); err != nil {
			return err
		}
	}
	return nil
}
