package mongodb

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/sisneve/rabbitmq-dashboard/internal/api/v1/profile"
	"github.com/sisneve/rabbitmq-dashboard/internal/database"
	"github.com/sisneve/rabbitmq-dashboard/internal/models"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type ACLHandler struct {
	Collection *mongo.Collection
	*CollectorHolder[models.ACL]
}

func NewACLHandler(collection *mongo.Collection) *ACLHandler {
	return &ACLHandler{
		Collection:      collection,
		CollectorHolder: NewCollectorHolder[models.ACL](collection),
	}
}

// interface guard clause.
var _ database.ACLRepository = (*ACLHandler)(nil)

var ErrACLNotFound = fmt.Errorf("acl not found. %w", database.ErrResourceNotFound)

// interface guard clause.
var _ profile.ACLStore = (*ACLHandler)(nil)

func (ah *ACLHandler) GetByGroupIDs(ctx context.Context, groupIDs []string) ([]*models.ACL, error) {
	filters := make([]database.Filter, len(groupIDs))
	for i, groupID := range groupIDs {
		filters[i] = database.NewFilter("group", database.FilterOpEqual, groupID)
	}

	cursor, err := ah.Collection.Find(ctx, bson.M{"$or": filters})
	if err != nil {
		return nil, fmt.Errorf("find acls for groups %v: %w", groupIDs, err)
	}
	defer func() {
		if err := cursor.Close(ctx); err != nil {
			slog.ErrorContext(ctx, "failed to close ACL cursor", "error", err)
		}
	}()

	var acls []*models.ACL
	if err := cursor.All(ctx, &acls); err != nil {
		return nil, fmt.Errorf("decode acls for groups %v: %w", groupIDs, err)
	}
	return acls, nil
}
