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
	Inialized   bool
}

type Collections struct {
	Alarms              *mongo.Collection
	Maintenance         *mongo.Collection
	MaintenanceEditLogs *mongo.Collection
	Notifications       *mongo.Collection
}

// MongoConstants
const (
	set       = "$set"
	id        = "_id"
	statusKey = "status"
)

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

	err = dbc.initCollections()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database collections. %w", err)
	}

	return &dbc, nil
}

func (dbc *Database) initCollections() error {

	client, err := mongo.Connect(
		options.Client().ApplyURI(dbc.uri),
		options.Client().SetTimeout(30*time.Second),
	)
	if err != nil {
		return err
	}

	dbc.Collections = &Collections{
		Alarms:              client.Database(dbc.db).Collection("alarms"),
		Maintenance:         client.Database(dbc.db).Collection("maintenance"),
		MaintenanceEditLogs: client.Database(dbc.db).Collection("maintenance_edit_logs"),
		Notifications:       client.Database(dbc.db).Collection("notifications"),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err = client.Ping(ctx, nil); err != nil {
		return fmt.Errorf("failed to verify connection to mongodb. %w", err)
	}

	dbc.Inialized = true

	return nil
}
