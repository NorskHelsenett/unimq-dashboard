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
	ACLs                *mongo.Collection
}

// MongoConstants
const (
	set       = "$set"
	id        = "_id"
	statusKey = "status"
)

type (
	databaseConfig struct {
		Host     string
		Port     int
		Username string
		Password string
		DB       string
	}

	databaseOptions func(*databaseConfig)
)

func newDatabaseConfig() *databaseConfig {
	return &databaseConfig{
		Host:     "localhost",
		Port:     27017,
		Username: "",
		Password: "",
		DB:       "unimq-dashboard",
	}
}

func WithHost(host string) databaseOptions {
	return func(dc *databaseConfig) {
		dc.Host = host
	}
}

func WithPort(port int) databaseOptions {
	return func(dc *databaseConfig) {
		dc.Port = port
	}
}

func WithUsername(username string) databaseOptions {
	return func(dc *databaseConfig) {
		dc.Username = username
	}
}

func WithPassword(password string) databaseOptions {
	return func(dc *databaseConfig) {
		dc.Password = password
	}
}

func WithDatabase(db string) databaseOptions {
	return func(dc *databaseConfig) {
		dc.DB = db
	}
}

func NewDatabase(opts ...databaseOptions) (*Database, error) {
	config := newDatabaseConfig()

	for _, opt := range opts {
		opt(config)
	}

	uri := fmt.Sprintf("mongodb://%s:%s@%s:%d",
		config.Username,
		config.Password,
		config.Host,
		config.Port,
	)

	client, err := mongo.Connect(
		options.Client().ApplyURI(uri),
		options.Client().SetTimeout(10*time.Second),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to mongodb at address %v. %w", uri, err)
	}

	dbc := Database{
		uri:         uri,
		db:          config.DB,
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
		ACLs:                client.Database(dbc.db).Collection("acls"),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err = client.Ping(ctx, nil); err != nil {
		return fmt.Errorf("failed to verify connection to mongodb. %w", err)
	}

	dbc.Inialized = true

	return nil
}
