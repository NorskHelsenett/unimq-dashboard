package rabbitmq

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConvertSingleFilterToPathParams(t *testing.T) {
	filter := NewFilter(ParameterName, FilterTypeVhost, "/")

	uri, filters, err := convertFiltersToPathParams([]Filter{filter}, "/vhost")

	require.NoError(t, err)
	assert.Equal(t, "/vhost/%2F", uri, "expected the name to be escaped into the path")
	assert.Empty(t, filters, "a single name filter is resolved by the path, so no post-filtering is needed")
}

func TestConvertMultipleFiltersToPathParams(t *testing.T) {
	vals := []string{"/", "test-vhost"}

	filters := make([]Filter, len(vals))
	for i, val := range vals {
		filters[i] = NewFilter(ParameterName, FilterTypeVhost, val)
	}

	uri, remainingFilters, err := convertFiltersToPathParams(filters, "/vhost")

	require.NoError(t, err)
	assert.Equal(t, "/vhost", uri, "multiple names cannot be expressed as a path parameter")
	assert.Equal(t, vals, remainingFilters, "expected every name to be returned for post-filtering")
}

func TestConvertFiltersToPathParamsRejectsUnsupportedParameter(t *testing.T) {
	cases := map[string][]Filter{
		"single filter": {
			NewFilter(ParameterPageSize, FilterTypeVhost, "10"),
		},
		"multiple filters": {
			NewFilter(ParameterName, FilterTypeVhost, "/"),
			NewFilter(ParameterPageSize, FilterTypeVhost, "10"),
		},
	}

	for name, filters := range cases {
		t.Run(name, func(t *testing.T) {
			uri, remaining, err := convertFiltersToPathParams(filters, "/vhost")

			require.ErrorIs(t, err, ErrUnsupportedFilterParameter)
			assert.Empty(t, uri)
			assert.Nil(t, remaining)
		})
	}
}

func TestConvertFiltersToQueryParams(t *testing.T) {
	cases := []struct {
		name        string
		filters     []Filter
		expectedURI string
	}{
		{
			name:        "no filters",
			filters:     nil,
			expectedURI: "",
		},
		{
			name:        "single filter",
			filters:     []Filter{NewFilter(ParameterName, FilterTypeVhost, "/")},
			expectedURI: "?name=%2F",
		},
		{
			name: "multiple filters",
			filters: []Filter{
				NewFilter(ParameterName, FilterTypeVhost, "/"),
				NewFilter(ParameterPageSize, FilterTypeVhost, "10"),
			},
			expectedURI: "?name=%2F&page_size=10",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expectedURI, convertFiltersToQueryParams(tc.filters))
		})
	}
}

func TestGetFiltersByType(t *testing.T) {
	vhostFilter := NewFilter(ParameterName, FilterTypeVhost, "/")
	queueFilter := NewFilter(ParameterName, FilterTypeQueue, "my-queue")
	filters := []Filter{vhostFilter, queueFilter}

	assert.Equal(t, []Filter{vhostFilter}, getFiltersByType(filters, FilterTypeVhost))
	assert.Equal(t, []Filter{queueFilter}, getFiltersByType(filters, FilterTypeQueue))
	assert.Empty(t, getFiltersByType(filters, FilterTypeChannel))
}
