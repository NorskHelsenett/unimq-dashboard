package storeservice

import "github.com/sisneve/rabbitmq-dashboard/internal/database"

type StoreService struct {
	Store database.Store
}

func NewStoreService(store database.Store) *StoreService {
	return &StoreService{
		Store: store,
	}
}
