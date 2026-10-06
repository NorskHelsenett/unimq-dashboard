package rabbitmq

import (
	"fmt"
	"net/url"
	"strings"
)

type (
	Filter struct {
		Parameter Parameter
		Type      FilterType
		Value     string
	}

	Parameter  string
	FilterType string
)

func NewFilter(parameter Parameter, filterType FilterType, value string) Filter {
	return Filter{
		Parameter: parameter,
		Type:      filterType,
		Value:     value,
	}
}

const (
	ParameterName              Parameter = "name"
	ParameterPage              Parameter = "page"
	ParameterPageSize          Parameter = "page_size"
	ParameterUseRegex          Parameter = "use_regex"
	ParameterPagination        Parameter = "pagination"
	ParameterEnableQueueTotals Parameter = "enable_queue_totals"
	ParmaeterDisableStats      Parameter = "disable_stats"

	FilterTypeVhost      FilterType = "vhost"
	FilterTypeQueue      FilterType = "queue"
	FilterTypeConnection FilterType = "connection"
	FilterTypeChannel    FilterType = "channel"
)

var (
	ErrUnsupportedFilterParameter = fmt.Errorf("unsupported filter parameter")
)

// Used where query parameters are supported.
func convertFiltersToQueryParams(filters []Filter) string {
	builder := &strings.Builder{}

	builder.WriteString("?")

	for i, filter := range filters {
		if i > 0 {
			builder.WriteString("&")
		}
		builder.WriteString(string(filter.Parameter))
		builder.WriteString("=")
		builder.WriteString(url.QueryEscape(filter.Value))
	}

	return builder.String()
}

// Convert filters to path parameters. Only supports the [ParameterName] filter.
// Returns the updated URI with the path parameter and a slice of filters that should be applied after fetching the resource(s).
// If the slice of filters is empty, it means that all resources should be fetched.
func convertFiltersToPathParams(filters []Filter, uri string) (string, []string, error) {
	var parameter string
	parameterFilter := make([]string, 0)

	switch {
	case len(filters) == 0:
		//  No Filters, fetch all.
	case len(filters) == 1:
		// One filter, check if it's the name filter, and use the path parameter to fetch the specific resource.
		if filters[0].Parameter == ParameterName {
			parameter = filters[0].Value
		} else {
			return "", nil, fmt.Errorf("%w: %s", ErrUnsupportedFilterParameter, filters[0].Parameter)
		}
	case len(filters) > 1:
		// Multiple filters, check if the name filter is present, query for all resources and filter the results based on the name filter.
		for _, filter := range filters {
			if filter.Parameter == ParameterName {
				parameterFilter = append(parameterFilter, filter.Value)
			} else {
				return "", nil, fmt.Errorf("%w: %s", ErrUnsupportedFilterParameter, filter.Parameter)
			}
		}
	default:
		return "", nil, fmt.Errorf("only one filter is supported for GetVhosts")
	}

	if parameter != "" {
		uri += "/" + parameter
	}

	return uri, parameterFilter, nil
}

func getFiltersByType(filters []Filter, filterType FilterType) []Filter {
	filtered := make([]Filter, 0)
	for _, filter := range filters {
		if filter.Type == filterType {
			filtered = append(filtered, filter)
		}
	}
	return filtered
}
