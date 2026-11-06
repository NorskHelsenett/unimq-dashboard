package storeservice

import "github.com/sisneve/rabbitmq-dashboard/internal/database"

type StoreService struct {
	database.Store
}

// NewStoreService creates a new StoreService instance with the provided database.Store implementation.
// TODO: Figure out if we need this layer or if we can just use the database.Store directly in the API handlers.
func NewStoreService(store database.Store) (*StoreService, error) {
	err := store.InitCollections()
	return &StoreService{
		store,
	}, err
}
