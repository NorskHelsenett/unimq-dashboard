package mongodb

import (
	"context"
	"fmt"

	"github.com/sisneve/rabbitmq-dashboard/internal/database"
	"github.com/sisneve/rabbitmq-dashboard/internal/models"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// TODO: This is a type used testing and development. It should not be used in production environments. It is a wrapper around the Database type that provides a Drop method for dropping entire collections.
type GeneratorDatabase struct {
	Database
}

// This type extends the CollectorHolder to include a Drop method that allows dropping the entire collection.
// This MUST NOT be used in production environments.
type CollectionHolderDropper[T any] struct {
	Collection *mongo.Collection
	*CollectorHolder[T]
}

func NewCollectionHolderDropper[T any](collection *mongo.Collection) *CollectionHolderDropper[T] {
	return &CollectionHolderDropper[T]{
		Collection:      collection,
		CollectorHolder: NewCollectorHolder[T](collection),
	}
}

func (c *CollectionHolderDropper[T]) Drop(ctx context.Context) error {
	if err := c.Collection.Drop(ctx); err != nil {
		return fmt.Errorf("failed to drop collection %s. %w", c.Collection.Name(), err)
	}
	return nil
}

type ACLDropHandler struct {
	Collection *mongo.Collection
	*CollectorHolder[models.ACL]
}

func (gdb *GeneratorDatabase) ACLHandler() database.ACLRepository {
	var out database.ACLRepository
	out = NewACLHandler(gdb.Collections.ACLs)
	return out
}
