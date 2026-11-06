// This package contains test functions and data for the database layer.
package testdatabase

import (
	"time"

	"github.com/sisneve/rabbitmq-dashboard/internal/clients/mongodb"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	MongoC       testcontainers.Container
	Username     = "testuser"
	Password     = "testpassword"
	DatabaseName = "testdb"
	Port         = 27017
)

func InitTestDatabaseContainer() {
	req := testcontainers.ContainerRequest{
		Image:        "mongo:latest",
		ExposedPorts: []string{"27017/tcp"},
		Env: map[string]string{
			"MONGO_INITDB_ROOT_USERNAME": Username,
			"MONGO_INITDB_ROOT_PASSWORD": Password,
			"MONGO_INITDB_DATABASE":      DatabaseName,
		},
		WaitingFor: wait.ForLog("Waiting for connections").WithStartupTimeout(60 * time.Second),
	}

	container, err := testcontainers.GenericContainer(
		nil,
		testcontainers.GenericContainerRequest{
			ContainerRequest: req,
			Started:          true,
		},
	)
	if err != nil {
		panic(err)
	}
	MongoC = container

}

func NewTestDatabase() *mongodb.Database {
	if MongoC == nil {
		panic("MongoDB container is not initialized. Call InitTestDatabaseContainer() first.")
	}
	db, err := mongodb.NewDatabase(
		mongodb.WithDatabase(DatabaseName),
		mongodb.WithUsername(Username),
		mongodb.WithPassword(Password),
		mongodb.WithPort(Port),
	)
	if err != nil {
		panic(err)
	}
	return db
}
