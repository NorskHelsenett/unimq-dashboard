package httpsuite_test

import (
	"testing"

	"github.com/sisneve/rabbitmq-dashboard/internal/api/httpsuite"
	"github.com/stretchr/testify/require"
)

type testCase struct {
	name        string
	request     testValidationRequest
	expectError bool
}

type testValidationRequest struct {
	Name string `validate:"required"`
	Age  int    `validate:"required,min=18"`
}

func TestValidateRequest(t *testing.T) {
	tests := []testCase{
		{
			name: "valid request",
			request: testValidationRequest{
				Name: "John Doe",
				Age:  25,
			},
			expectError: false,
		},
		{
			name: "missing name",
			request: testValidationRequest{
				Name: "",
				Age:  25,
			},
			expectError: true,
		},
		{
			name: "age below minimum",
			request: testValidationRequest{
				Name: "John Doe",
				Age:  17,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := httpsuite.IsRequestValid(tt.request)
			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
