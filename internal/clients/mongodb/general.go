package mongodb

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/sisneve/rabbitmq-dashboard/internal/database"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// This type generalizes the CRUD operations for any resource type T that implements the Collectorer interface.
// It provides methods to read, create, update, and delete resources in a MongoDB collection.
// To use this type, embed it in a struct that implements the correct model type and pass the appropriate MongoDB collection to the constructor.
type CollectorHolder[T any] struct {
	Collection *mongo.Collection
}

func NewCollectorHolder[T any](collection *mongo.Collection) *CollectorHolder[T] {
	return &CollectorHolder[T]{
		Collection: collection,
	}
}

func (c *CollectorHolder[T]) Read(ctx context.Context, filters ...database.Filter) ([]*T, error) {
	filter := bson.M{}
	for _, f := range filters {
		filter[f.Property] = f.Value
	}

	cursor, err := c.Collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to find resource(s). %w", err)
	}
	defer func() {
		if err := cursor.Close(ctx); err != nil {
			slog.ErrorContext(ctx, "failed to close resource cursor", "error", err)
		}
	}()

	var resource []*T
	if err := cursor.All(ctx, &resource); err != nil {
		return nil, fmt.Errorf("failed to decode resource(s). %w", err)
	}
	return resource, nil
}

func (c *CollectorHolder[T]) Create(ctx context.Context, resources []*T) error {
	result, err := c.Collection.InsertMany(ctx, resources)
	if err != nil {
		return fmt.Errorf("failed to create resource(s) %v. %w", resources, err)
	}
	if len(result.InsertedIDs) != len(resources) {
		return fmt.Errorf("failed to create all resources. expected %d, got %d", len(resources), len(result.InsertedIDs))
	}

	return nil
}

func (c *CollectorHolder[T]) Update(ctx context.Context, resource *T, filters ...database.Filter) error {
	filter := parseFilters(filters...)

	result, err := c.Collection.ReplaceOne(ctx, filter, resource)
	if err != nil {
		return fmt.Errorf("failed to update resource %v. %w", id, err)
	}
	if result.MatchedCount == 0 {
		return database.ErrResourceNotFound
	}
	return nil
}

func (c *CollectorHolder[T]) Delete(ctx context.Context, filters ...database.Filter) error {
	filter := parseFilters(filters...)

	result, err := c.Collection.DeleteMany(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to delete resource(s) %v. %w", filter, err)
	}
	if result.DeletedCount == 0 {
		return database.ErrResourceNotFound
	}
	return nil
}

func parseFilters(filters ...database.Filter) bson.M {
	filter := bson.M{}
	for _, f := range filters {
		filter[f.Property] = f.Value
	}
	return filter
}

func (c *CollectorHolder[T]) Patch(ctx context.Context, patch map[string]any, filters ...database.Filter) error {
	filter := parseFilters(filters...)
	update := bson.M{"$set": patch}

	result, err := c.Collection.UpdateMany(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to patch resource %v. %w", id, err)
	}
	if result.MatchedCount == 0 {
		return database.ErrResourceNotFound
	}
	return nil
}

func (c *CollectorHolder[T]) Put(ctx context.Context, put *T, filters ...database.Filter) error {
	filter := parseFilters(filters...)
	update := bson.M{"$set": put}
	opts := options.UpdateOne().SetUpsert(true)

	_, err := c.Collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return fmt.Errorf("put resource %v failed. %w", put, err)
	}
	return nil
}
