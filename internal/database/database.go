package database

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Database struct {
	uri         string
	db          string
	client      *mongo.Client
	Collections *Collections
}

type Collections struct {
	Alarms        *mongo.Collection
	Maintenance   *mongo.Collection
	Notifications *mongo.Collection
}

func BuildURI(host, username, password string, port int) string {
	return fmt.Sprintf("mongodb://%s:%s@%s:%d",
		username,
		password,
		host,
		port,
	)
}

func NewDatabase(uri, db string) (*Database, error) {

	client, err := mongo.Connect(
		options.Client().ApplyURI(uri),
		options.Client().SetTimeout(10*time.Second),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to mongodb at address %v. %w", uri, err)
	}

	dbc := Database{
		uri:         uri,
		db:          db,
		client:      client,
		Collections: nil,
	}

	return &dbc, nil
}

func (dbc *Database) InitCollections() error {
	client, err := mongo.Connect(
		options.Client().ApplyURI(dbc.uri),
		options.Client().SetTimeout(10*time.Second),
	)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	dbc.Collections = &Collections{
		Alarms:        client.Database(dbc.db).Collection("alarms"),
		Maintenance:   client.Database(dbc.db).Collection("maintenance"),
		Notifications: client.Database(dbc.db).Collection("notifications"),
	}

	if err = client.Ping(ctx, nil); err != nil {
		return fmt.Errorf("failed to verify connection to mongodb. %w", err)
	}

	return nil
}
