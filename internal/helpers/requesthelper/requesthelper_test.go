package requesthelper_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/sisneve/rabbitmq-dashboard/internal/helpers/requesthelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newRequestWithParam builds an *http.Request carrying a chi RouteContext
// with the given URL param set, mimicking how chi populates params when
// routing a real request.
func newRequestWithParam(key, value string) *http.Request {
	rctx := chi.NewRouteContext()
	if value != "" {
		rctx.URLParams.Add(key, value)
	}
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
}

type paramReader struct {
	name     string
	paramKey string
	read     func(*http.Request) (string, error)
}

// readers covers every simple "required URL param, no unescaping quirks"
// reader.
var readers = []paramReader{
	{"ReadVhostFromRequest", "vhost-name", requesthelper.ReadVhostFromRequest},
	{"ReadRuleIDFromRequest", "rule-id", requesthelper.ReadRuleIDFromRequest},
	{"ReadRecipientIDFromRequest", "recipient-id", requesthelper.ReadRecipientIDFromRequest},
	{"ReadMaintenanceIDFromRequest", "maintenance-id", requesthelper.ReadMaintenanceIDFromRequest},
	{"ReadQueueIDFromRequest", "queue-id", requesthelper.ReadQueueIDFromRequest},
}

func TestParamReaders_Valid(t *testing.T) {
	for _, r := range readers {
		t.Run(r.name, func(t *testing.T) {
			req := newRequestWithParam(r.paramKey, "some-value")
			got, err := r.read(req)
			require.NoError(t, err)
			assert.Equal(t, "some-value", got)
		})
	}
}

func TestParamReaders_Missing(t *testing.T) {
	for _, r := range readers {
		t.Run(r.name, func(t *testing.T) {
			req := newRequestWithParam(r.paramKey, "")
			_, err := r.read(req)
			require.ErrorIs(t, err, requesthelper.ErrMissingRequiredParameter)
			assert.Contains(t, err.Error(), r.paramKey)
		})
	}
}

// unescapingReaders covers readers that perform URL unescaping, which has some quirks
// that we want to test explicitly.
var unescapingReaders = []paramReader{
	{"ReadVhostFromRequest", "vhost-name", requesthelper.ReadVhostFromRequest},
	{"ReadQueueIDFromRequest", "queue-id", requesthelper.ReadQueueIDFromRequest},
}

func TestParamReaders_Unescaping(t *testing.T) {
	for _, r := range unescapingReaders {
		t.Run(r.name, func(t *testing.T) {
			t.Run("percent-encoded characters are decoded", func(t *testing.T) {
				req := newRequestWithParam(r.paramKey, "my%20value")
				got, err := r.read(req)
				require.NoError(t, err)
				assert.Equal(t, "my value", got)
			})

			t.Run("plus sign is preserved, not turned into a space", func(t *testing.T) {
				req := newRequestWithParam(r.paramKey, "foo+bar")
				got, err := r.read(req)
				require.NoError(t, err)
				assert.Equal(t, "foo+bar", got)
			})

			t.Run("invalid percent-encoding fails to decode", func(t *testing.T) {
				req := newRequestWithParam(r.paramKey, "%zz")
				_, err := r.read(req)
				require.ErrorIs(t, err, requesthelper.ErrFailedToDecodeParameter)
			})
		})
	}
}

func TestReadVhostFromRequest_PercentEncodedSlash(t *testing.T) {
	req := newRequestWithParam("vhost-name", "%2F")
	vhost, err := requesthelper.ReadVhostFromRequest(req)
	require.NoError(t, err)
	assert.Equal(t, "/", vhost)
}
