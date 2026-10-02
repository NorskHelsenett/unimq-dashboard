package database

import (
	"context"
	"fmt"
	"net/url"
	"strings"
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
	TimeoutSecs int
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
		Host        string
		Port        int
		Username    string
		Password    string
		DB          string
		TimeoutSecs int
	}

	databaseOptions func(*databaseConfig)
)

func newDatabaseConfig() *databaseConfig {
	return &databaseConfig{
		Host:        "localhost",
		Port:        27017,
		Username:    "",
		Password:    "",
		DB:          "unimq-dashboard",
		TimeoutSecs: 30,
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

func WithTimeout(timeoutSecs int) databaseOptions {
	return func(dc *databaseConfig) {
		dc.TimeoutSecs = timeoutSecs
	}
}

func NewDatabase(opts ...databaseOptions) (*Database, error) {
	config := newDatabaseConfig()

	for _, opt := range opts {
		opt(config)
	}

	uri := CreateUri(config.Host, config.Port, config.Username, config.Password)

	dbc := Database{
		uri:         uri,
		db:          config.DB,
		client:      nil,
		Collections: nil,
		Inialized:   false,
		TimeoutSecs: config.TimeoutSecs,
	}

	client, err := CreateClient(uri, config.TimeoutSecs)
	if err != nil {
		return nil, fmt.Errorf("failed to create database client. %w", err)
	}
	dbc.client = client

	err = dbc.initCollections()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database collections. %w", err)
	}

	return &dbc, nil
}

func (dbc *Database) Close(timeoutSecs int) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutSecs)*time.Second)
	defer cancel()

	if err := dbc.client.Disconnect(ctx); err != nil {
		return fmt.Errorf("failed to disconnect from database. %w", err)
	}

	return nil
}

func (dbc *Database) Ping(ctx context.Context, timeoutSecs int) error {
	ctx, cancel := context.WithTimeout(ctx, time.Duration(timeoutSecs)*time.Second)
	defer cancel()

	if err := dbc.client.Ping(ctx, nil); err != nil {
		return fmt.Errorf("failed to ping database. %w", err)
	}

	return nil
}

func CreateUri(host string, port int, username, password string) string {

	// Remove any existing "mongodb://" prefix from the host string
	host = strings.TrimPrefix(host, "mongodb://")

	if username == "" && password == "" {
		return fmt.Sprintf("mongodb://%s:%d", url.QueryEscape(host), port)
	}

	userInfo := url.UserPassword(username, password)

	return fmt.Sprintf("mongodb://%s@%s:%d",
		userInfo.String(),
		url.PathEscape(host),
		port,
	)
}

func CreateClient(uri string, timeoutSecs int) (*mongo.Client, error) {
	client, err := mongo.Connect(
		options.Client().ApplyURI(uri),
		options.Client().SetTimeout(time.Duration(timeoutSecs)*time.Second),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to mongodb. %w", err)
	}

	return client, nil
}

func (dbc *Database) initCollections() error {

	dbc.Collections = &Collections{
		Alarms:              dbc.client.Database(dbc.db).Collection("alarms"),
		Maintenance:         dbc.client.Database(dbc.db).Collection("maintenance"),
		MaintenanceEditLogs: dbc.client.Database(dbc.db).Collection("maintenance_edit_logs"),
		Notifications:       dbc.client.Database(dbc.db).Collection("notifications"),
		ACLs:                dbc.client.Database(dbc.db).Collection("acls"),
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(dbc.TimeoutSecs)*time.Second)
	defer cancel()

	if err := dbc.client.Ping(ctx, nil); err != nil {
		return fmt.Errorf("failed to verify connection to mongodb. %w", err)
	}

	dbc.Inialized = true

	return nil
}
