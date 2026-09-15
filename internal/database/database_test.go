package database_test

import (
	"testing"

	"github.com/sisneve/rabbitmq-dashboard/internal/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testCasesConnectionStrings struct {
	name          string
	host          string
	port          int
	username      string
	password      string
	expectedUri   string
	expectedError error
}

func TestDatabaseConnection(t *testing.T) {

	testcases := []testCasesConnectionStrings{
		{
			name:          "Default connection string",
			host:          "localhost",
			port:          27017,
			username:      "",
			password:      "",
			expectedUri:   "mongodb://localhost:27017",
			expectedError: nil,
		},
		{
			name:          "Connection string with username and password",
			host:          "localhost",
			port:          27017,
			username:      "user",
			password:      "pass",
			expectedUri:   "mongodb://user:pass@localhost:27017",
			expectedError: nil,
		},
		{
			name:          "Connection string with special characters in username and password",
			host:          "localhost",
			port:          27017,
			username:      "user@name",
			password:      "p@ssw0rd",
			expectedUri:   "mongodb://user%40name:p%40ssw0rd@localhost:27017",
			expectedError: nil,
		},
		{
			name:          "Connection string with invalid host",
			host:          "invalidhost",
			port:          27017,
			username:      "",
			password:      "",
			expectedUri:   "mongodb://invalidhost:27017",
			expectedError: nil,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			uri := database.CreateUri(tc.host, tc.port, tc.username, tc.password)

			require.Equalf(t, tc.expectedUri, uri, "Expected URI: %v, but got: %v", tc.expectedUri, uri)

			client, err := database.CreateClient(uri, 30)

			require.Equalf(t, tc.expectedError, err, "Expected error: %v, but got: %v of ", tc.expectedError, err)

			assert.NotNilf(t, client, "Expected client to be not nil, but got nil")
		})
	}

}
