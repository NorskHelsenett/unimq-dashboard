package requesthelper

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"
)

var (
	ErrMissingRequiredParameter = errors.New("missing required parameter")
	ErrFailedToDecodeParameter  = errors.New("failed to decode parameter")
)

func ReadVhostFromRequest(r *http.Request) (string, error) {
	vhost := chi.URLParam(r, "vhost-name")
	if vhost == "" {
		return "", fmt.Errorf("%w: vhost-name", ErrMissingRequiredParameter)
	}

	eVhost, err := url.PathUnescape(vhost)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrFailedToDecodeParameter, err)
	}

	return eVhost, nil
}

func ReadRuleIDFromRequest(r *http.Request) (string, error) {
	id := chi.URLParam(r, "rule-id")
	if id == "" {
		return "", fmt.Errorf("%w: rule-id", ErrMissingRequiredParameter)
	}

	return id, nil
}

func ReadRecipientIDFromRequest(r *http.Request) (string, error) {
	id := chi.URLParam(r, "recipient-id")
	if id == "" {
		return "", fmt.Errorf("%w: recipient-id", ErrMissingRequiredParameter)
	}

	return id, nil
}

func ReadMaintenanceIDFromRequest(r *http.Request) (string, error) {
	id := chi.URLParam(r, "maintenance-id")
	if id == "" {
		return "", fmt.Errorf("%w: maintenance-id", ErrMissingRequiredParameter)
	}

	return id, nil
}
