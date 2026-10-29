package mongodb

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/sisneve/rabbitmq-dashboard/internal/database"
	"github.com/sisneve/rabbitmq-dashboard/internal/models"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type ACLHandler struct {
	Collection *mongo.Collection
}

// interface guard clause
var _ database.ACLRepository = (*ACLHandler)(nil)

var ErrACLNotFound = fmt.Errorf("acl not found. %w", database.ErrResourceNotFound)

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

func (ah *ACLHandler) Read(ctx context.Context, filters ...database.Filter) ([]*models.ACL, error) {
	cursor, err := ah.Collection.Find(ctx, bson.D{})
	if err != nil {
		return nil, fmt.Errorf("find acls: %w", err)
	}
	defer func() {
		if err := cursor.Close(ctx); err != nil {
			slog.ErrorContext(ctx, "failed to close ACL cursor", "error", err)
		}
	}()

	var acls []*models.ACL
	if err := cursor.All(ctx, &acls); err != nil {
		return nil, fmt.Errorf("decode acls: %w", err)
	}
	return acls, nil
}

func (ah *ACLHandler) GetByID(ctx context.Context, id string) (*models.ACL, error) {
	var acl models.ACL
	if err := ah.Collection.FindOne(ctx, bson.M{"id": id}).Decode(&acl); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrACLNotFound
		}
		return nil, fmt.Errorf("find acl for group %q: %w", id, err)
	}
	return &acl, nil
}

func (ah *ACLHandler) Create(ctx context.Context, acl *models.ACL) error {
	if err := acl.Validate(); err != nil {
		return err
	}

	_, err := ah.Collection.InsertOne(ctx, acl)
	if err != nil {
		return fmt.Errorf("insert acl for group %q: %w", acl.Group, err)
	}
	return nil
}

func (ah *ACLHandler) Update(ctx context.Context, acl *models.ACL) error {
	if err := acl.Validate(); err != nil {
		return err
	}

	result, err := ah.Collection.ReplaceOne(ctx, bson.M{"id": acl.Group}, acl)
	if err != nil {
		return fmt.Errorf("replace acl for group %q: %w", acl.Group, err)
	}
	if result.MatchedCount == 0 {
		return ErrACLNotFound
	}
	return nil
}

func (ah *ACLHandler) Delete(ctx context.Context, filters ...database.Filter) error {
	filter := bson.M{}
	for _, f := range filters {
		filter[f.Property] = f.Value
	}
	result, err := ah.Collection.DeleteMany(ctx, filter)
	if err != nil {
		return fmt.Errorf("delete acl for group %q: %w", id, err)
	}
	if result.DeletedCount == 0 {
		return ErrACLNotFound
	}
	return nil
}

func (ah *ACLHandler) Patch(ctx context.Context, id string, updates map[string]any) error {
	update := bson.M{"$set": updates}
	result, err := ah.Collection.UpdateOne(ctx, bson.M{"id": id}, update)
	if err != nil {
		return fmt.Errorf("update acl for group %q: %w", id, err)
	}
	if result.MatchedCount == 0 {
		return ErrACLNotFound
	}
	return nil
}

func (ah *ACLHandler) Put(ctx context.Context, acl *models.ACL) error {
	if err := acl.Validate(); err != nil {
		return err
	}

	filter := bson.M{"id": acl.Group}
	update := bson.M{"$set": acl}
	opts := options.UpdateOne().SetUpsert(true)

	_, err := ah.Collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return fmt.Errorf("upsert acl for group %q: %w", acl.Group, err)
	}
	return nil
}

func (ah *ACLHandler) GetByGroupID(ctx context.Context, groupID string) (*models.ACL, error) {
	var acl models.ACL
	if err := ah.Collection.FindOne(ctx, bson.M{"group": groupID}).Decode(&acl); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrACLNotFound
		}
		return nil, fmt.Errorf("find acl for group %q: %w", groupID, err)
	}
	return &acl, nil
}
