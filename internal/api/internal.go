package api

import (
	"context"
	"errors"
	"fmt"

	"github.com/sisneve/rabbitmq-dashboard/internal/models"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// Internal is for functionality not exposed to the API.
// Typically helper functions that use multiple API endpoints.

var (
	errFailedToCreateNotificationHost = errors.New("failed to create notification host")
)

func (rc *APIService) ensureNotificationHostExists(ctx context.Context, vhost string) (*models.VhostNotification, error) {

	vhostNotification, err := rc.DB.GetNotification(ctx, vhost)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			vhostNotification = &models.VhostNotification{
				Name:       vhost,
				Rules:      []*models.AlarmRule{},
				Recipients: []*models.Recipient{},
				Notified:   false,
			}
			err := rc.DB.AddNotification(ctx, *vhostNotification)
			if err != nil {
				return nil, fmt.Errorf("%w. %w", errFailedToCreateNotificationHost, err)
			}
		}
	}

	if vhostNotification == nil {
		vhostNotification, err = rc.DB.GetVhost(ctx, vhost)
		if err != nil {
			return nil, fmt.Errorf("%w. %w", errFailedToCreateNotificationHost, err)
		}
	}

	return vhostNotification, nil
}
