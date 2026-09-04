package database

import (
	"context"
	"fmt"
	"log/slog"
	"slices"
	"time"
	"uuid"

	"github.com/sisneve/rabbitmq-dashboard/internal/models"
)

const (
	recipientID      = "e45957ef-b817-414e-95e8-c4ea89c4ad3e"
	recipientEmailID = "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
)

var (
	alarmsSeed = []models.AlarmRule{
		{
			ID:        "",
			QueueName: "",
			Name:      "High queue message size",
			Threshold: 50.0,
			Type:      models.AlarmTypeQueueMessages,
			Enabled:   true,
			Status:    models.AlarmStatusActive,
			LastFired: new(time.Now().Add(-10 * time.Minute)),
			LastValue: new(47.0),
		},
		{
			ID:        "",
			QueueName: "",
			Name:      "High queue size",
			Threshold: 15.0,
			Type:      models.AlarmTypeQueueSize,
			Enabled:   true,
			Status:    models.AlarmStatusActive,
			LastFired: new(time.Now().Add(-10 * time.Minute)),
			LastValue: new(15.0),
		},
		{
			ID:        "",
			QueueName: "",
			Name:      "High channel count",
			Threshold: 10.0,
			Type:      models.AlarmTypeChannels,
			Enabled:   true,
			Status:    models.AlarmStatusFiring,
			LastFired: new(time.Now().Add(-10 * time.Minute)),
			LastValue: new(47.0),
		},
		{
			ID:        "",
			QueueName: "",
			Name:      "High connection count",
			Threshold: 10.0,
			Type:      models.AlarmTypeConnections,
			Enabled:   true,
			Status:    models.AlarmStatusFiring,
			LastFired: new(time.Now().Add(-10 * time.Minute)),
			LastValue: new(47.0),
		},
	}

	maintenanceEntries = []models.MaintenanceEntry{
		{
			ID:          "e45957ef-b817-414e-95e8-c4ea89c4ad3e",
			Description: "Test maintenance entry 1",
			Start:       time.Now(),
			End:         time.Now().Add(2 * time.Hour),
			Status:      models.MaintenanceStatusScheduled,
		},
		{
			ID:          "1fdbdeba-13c8-496f-9747-51a40679034e",
			Description: "Test maintenance entry 2",
			Start:       time.Now(),
			End:         time.Now().Add(2 * time.Hour),
			Status:      models.MaintenanceStatusInProgress,
		},
		{
			ID:          "2fdbdeba-13c8-496f-9747-51a40679034e",
			Description: "Test maintenance entry 3",
			Start:       time.Now(),
			End:         time.Now().Add(2 * time.Hour),
			Status:      models.MaintenanceStatusDone,
		},
	}
)

// Seed populates the database with initial data for testing and development purposes.
// It drops existing collections and inserts predefined notification rules,
// recipients, alarms, maintenance entries, and maintenance logs based on the provided queue mapping.
// The provided queue mapping is a map where the keys are vhost names and the values are slices of queue names associated with each vhost.
func (dbc *Database) Seed(ctx context.Context, queueMapping map[string][]string) error {

	if err := dbc.Collections.Notifications.Drop(ctx); err != nil {
		return fmt.Errorf("failed to drop notifications collection: %w", err)
	}

	if err := dbc.Collections.Alarms.Drop(ctx); err != nil {
		return fmt.Errorf("failed to drop alarms collection: %w", err)
	}

	if err := dbc.Collections.Maintenance.Drop(ctx); err != nil {
		return fmt.Errorf("failed to drop maintenance collection: %w", err)
	}

	if err := dbc.Collections.MaintenanceEditLogs.Drop(ctx); err != nil {
		return fmt.Errorf("failed to drop maintenance logs collection: %w", err)
	}

	vhosts := make([]string, 0, len(queueMapping))
	for vhost := range queueMapping {
		vhosts = append(vhosts, vhost)
	}

	for _, vhost := range vhosts {

		queues := queueMapping[vhost]
		alarmSeedsCount := len(alarmsSeed)
		alarmRules := make([]models.AlarmRule, 0, alarmSeedsCount)

		for _, queue := range queues {
			alarmRules = slices.Clone(alarmsSeed)
			for i := range alarmRules {
				alarmRules[i].ID = uuid.New().String()
				alarmRules[i].QueueName = queue
			}
		}

		err := dbc.seedNotifications(ctx, vhost, alarmRules)
		if err != nil {
			return err
		}

		err = dbc.SeedAlarms(ctx, vhost, alarmRules)
		if err != nil {
			return err
		}
	}

	err := dbc.seedMaintenace(ctx)
	if err != nil {
		return err
	}

	err = dbc.SeedMaintenanceLogs(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (dbc *Database) seedNotifications(ctx context.Context, name string, alarmRules []models.AlarmRule) error {

	err := dbc.EnsureVhostExists(ctx, name)
	if err != nil {
		return err
	}

	err = dbc.seedNotificationRecipients(ctx, name)
	if err != nil {
		return err
	}

	err = dbc.seedNotificationRules(ctx, name, alarmRules)
	if err != nil {
		return err
	}

	return nil
}

func (dbc *Database) seedNotificationRecipients(ctx context.Context, name string) error {

	slog.DebugContext(ctx, "seeding notification recipient", "vhost", name, "recipientID", recipientID)

	err := dbc.AddNotificationRecipient(ctx, name, &models.Recipient{
		ID:   recipientID,
		Name: "Matias",
		URL:  "https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX",
		Type: models.RecipientTypeWebhook,
	})
	if err != nil {
		return err
	}

	err = dbc.AddNotificationRecipient(ctx, name, &models.Recipient{
		ID:    recipientEmailID,
		Name:  "Matias",
		Email: "matias.nordmann@example.no",
		Type:  models.RecipientTypeEmail,
	})
	if err != nil {
		return err
	}

	return nil
}

func (dbc *Database) seedNotificationRules(ctx context.Context, vhost string, alarmRules []models.AlarmRule) error {

	for _, missingRule := range alarmRules {
		err := dbc.AddNotificationRule(ctx, vhost, &missingRule)
		if err != nil {
			return err
		}
	}

	return nil
}

func (dbc *Database) seedMaintenace(ctx context.Context) error {

	for _, MaintenanceEntry := range maintenanceEntries {
		err := dbc.AddMaintenanceEntry(ctx, &MaintenanceEntry)
		if err != nil {
			return err
		}
	}
	return nil
}

func (dbc *Database) SeedMaintenanceEntry(ctx context.Context, entry *models.MaintenanceEntry) error {

	err := dbc.AddMaintenanceEntry(ctx, entry)
	if err != nil {
		return err
	}

	return nil
}

func (dbc *Database) SeedMaintenanceLogs(ctx context.Context) error {

	for _, MaintenanceEntry := range maintenanceEntries {

		logEntries := []models.MaintenanceEditLog{}
		logEntries = append(logEntries, *models.NewMaintenaceEditLog(MaintenanceEntry.ID, "Test maintenance log entry 1", time.Now(), time.Now().Add(2*time.Hour), "Initial creation", "user1"))
		logEntries = append(logEntries, *models.NewMaintenaceEditLog(MaintenanceEntry.ID, "Test maintenance log entry 2", time.Now(), time.Now().Add(2*time.Hour), "Updated description", "user2"))
		logEntries = append(logEntries, *models.NewMaintenaceEditLog(MaintenanceEntry.ID, "Test maintenance log entry 3", time.Now(), time.Now().Add(2*time.Hour), "Updated start and end times", "user3"))

		for _, log := range logEntries {
			err := dbc.AddMaintenanceEditLog(ctx, &log)
			if err != nil {
				return err
			}
		}

	}

	return nil
}

func (dbc *Database) SeedAlarms(ctx context.Context, name string, alarmRules []models.AlarmRule) error {

	for _, rule := range alarmRules {

		entries := models.AlarmEntry{
			AlarmID: rule.ID,
			Entries: []models.LogEntry{},
		}

		val := 5.0
		entries.Entries = append(entries.Entries, models.NewLogEntry(models.LogEventFired, &val, 3.4, rule.Type))
		entries.Entries = append(entries.Entries, models.NewLogEntry(models.LogEventResolved, &val, 42.0, rule.Type))
		entries.Entries = append(entries.Entries, models.NewLogEntry(models.LogEventFired, &val, 0.67, rule.Type))
		entries.Entries = append(entries.Entries, models.NewLogEntry(models.LogEventResolved, &val, 69.0, rule.Type))
		entries.Entries = append(entries.Entries, models.NewLogEntry(models.LogEventFired, new(47.0), 10.0, rule.Type))

		err := dbc.AddAlarm(ctx, &entries)
		if err != nil {
			return err
		}
	}

	return nil
}
