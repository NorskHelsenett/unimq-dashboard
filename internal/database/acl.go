package database

import (
	"context"
	"errors"
	"fmt"

	"github.com/sisneve/rabbitmq-dashboard/internal/models"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var ErrACLNotFound = errors.New("acl not found")

func (dbc *Database) GetACLs(ctx context.Context) ([]models.ACL, error) {
	cursor, err := dbc.Collections.ACLs.Find(ctx, bson.D{})
	if err != nil {
		return nil, fmt.Errorf("find acls: %w", err)
	}
	defer cursor.Close(ctx)

	var acls []models.ACL
	if err := cursor.All(ctx, &acls); err != nil {
		return nil, fmt.Errorf("decode acls: %w", err)
	}
	return acls, nil
}

func (dbc *Database) GetACLsForGroups(ctx context.Context, groups []string) ([]models.ACL, error) {
	if len(groups) == 0 {
		return []models.ACL{}, nil
	}

	cursor, err := dbc.Collections.ACLs.Find(ctx, bson.M{id: bson.M{"$in": groups}})
	if err != nil {
		return nil, fmt.Errorf("find acls for groups: %w", err)
	}
	defer cursor.Close(ctx)

	var acls []models.ACL
	if err := cursor.All(ctx, &acls); err != nil {
		return nil, fmt.Errorf("decode acls for groups: %w", err)
	}
	return acls, nil
}

func (dbc *Database) GetACL(ctx context.Context, group string) (*models.ACL, error) {
	var acl models.ACL
	if err := dbc.Collections.ACLs.FindOne(ctx, bson.M{id: group}).Decode(&acl); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrACLNotFound
		}
		return nil, fmt.Errorf("find acl for group %q: %w", group, err)
	}
	return &acl, nil
}

func (dbc *Database) UpsertACL(ctx context.Context, acl *models.ACL) error {
	if err := acl.Validate(); err != nil {
		return err
	}

	_, err := dbc.Collections.ACLs.ReplaceOne(
		ctx,
		bson.M{id: acl.Group},
		acl,
		options.Replace().SetUpsert(true),
	)
	if err != nil {
		return fmt.Errorf("upsert acl for group %q: %w", acl.Group, err)
	}
	return nil
}

func (dbc *Database) DeleteACL(ctx context.Context, group string) error {
	result, err := dbc.Collections.ACLs.DeleteOne(ctx, bson.M{id: group})
	if err != nil {
		return fmt.Errorf("delete acl for group %q: %w", group, err)
	}
	if result.DeletedCount == 0 {
		return ErrACLNotFound
	}
	return nil
}
