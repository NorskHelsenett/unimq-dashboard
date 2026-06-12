package database

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/sisneve/rabbitmq-dashboard/internal/models"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func (dbc *Database) Seed(ctx context.Context) error {
	// Seed notifications
	err := dbc.seedNotifications(ctx)
	if err != nil {
		return err
	}

	err = dbc.seedMaintenace(ctx)
	if err != nil {
		return err
	}

	err = dbc.SeedAlarms(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (dbc *Database) seedNotifications(ctx context.Context) error {

	name := "test-Name"
	err := dbc.EnsureVhostExists(ctx, name)
	if err != nil {
		return err
	}

	err = dbc.seedNotificationRecipients(ctx, name)
	if err != nil {
		return err
	}

	err = dbc.seedNotificationRules(ctx, name)
	if err != nil {
		return err
	}

	return nil
}

func (dbc *Database) seedNotificationRecipients(ctx context.Context, name string) error {

	recipientID := "test-recipient"
	recipient, err := dbc.GetNotificationRecipient(ctx, name, recipientID)
	if err == nil {
		if recipient == nil {
			return fmt.Errorf("unexpected nil recipient")
		}
		if recipientID != recipient.ID {
			return fmt.Errorf("unexpected recipient ID. expected %s, got %s", recipientID, recipient.ID)
		}
		return nil
	}

	if !errors.Is(err, mongo.ErrNoDocuments) && !errors.Is(err, ErrRecipientNotFound) {
		return fmt.Errorf("failed to check existing recipient entry. %w", err)
	}

	err = dbc.AddNotificationRecipient(ctx, name, models.Recipient{
		ID:   recipientID,
		Name: "Matias",
		Type: "email",
	})
	if err != nil {
		return err
	}

	return nil
}

func (dbc *Database) seedNotificationRules(ctx context.Context, name string) error {

	ruleID := "test-rule"
	rules, err := dbc.GetNotificationRules(ctx, name)
	if err != nil {
		if !errors.Is(err, mongo.ErrNoDocuments) {
			return fmt.Errorf("failed to get existing notification rules. %w", err)
		}
	}

	if len(rules) != 0 {
		for _, r := range rules {
			if r.ID == ruleID {
				return nil
			}
		}
	}

	err = dbc.AddNotificationRule(ctx, name, models.AlarmRule{
		ID:        ruleID,
		Name:      "forks",
		Threshold: 10.0,
		Type:      "equal",
		Message:   "Test rule triggered",
		Enabled:   true,
		Status:    "active",
	})
	if err != nil {
		return err
	}

	return nil
}

func (dbc *Database) seedMaintenace(ctx context.Context) error {

	name := "test-maintenance"
	existing, err := dbc.GetMaintenanceEntry(ctx, name)
	if err != nil {
		if !errors.Is(err, mongo.ErrNoDocuments) {
			return fmt.Errorf("failed to check existing maintenance entry. %w", err)
		}
	} else {
		if existing == nil {
			return fmt.Errorf("unexpected nil maintenance entry")
		}
		if existing.ID == name {
			return nil
		}
		return fmt.Errorf("unexpected maintenance entry ID. expected %s, got %s", "test-maintenance", existing.ID)
	}

	maintenance := models.MaintenanceEntry{
		ID:          name,
		Description: "Test maintenance entry",
		Start:       time.Now(),
		End:         time.Now().Add(2 * time.Hour),
		Status:      models.MaintenanceStatusScheduled,
	}

	err = dbc.AddMaintenanceEntry(ctx, &maintenance)
	if err != nil {
		return err
	}
	return nil
}

func (dbc *Database) SeedAlarms(ctx context.Context) error {
	alarmID := "test-alarm"
	alarm, err := dbc.GetAlarm(ctx, alarmID)
	if err == nil {
		if alarm == nil {
			return fmt.Errorf("unexpected nil alarm entry")
		}
		if alarm.AlarmID != alarmID {
			return fmt.Errorf("unexpected alarm entry name. expected %s, got %s", alarmID, alarm.AlarmID)
		}
		return nil
	}
	if !errors.Is(err, mongo.ErrNoDocuments) {
		return fmt.Errorf("failed to check existing alarm entry. %w", err)
	}

	entry := models.AlarmEntry{
		AlarmID: alarmID,
		Entries: []models.LogEntry{
			{
				Timestamp: time.Now(),
				Event:     models.LogEvent("Test alarm triggered"),
				Value:     nil,
				Threshold: 0.0,
			},
			{
				Timestamp: time.Now().Add(1 * time.Hour),
				Event:     models.LogEvent("Test alarm resolved"),
				Value:     nil,
				Threshold: 0.0,
			},
		},
	}

	err = dbc.AddAlarm(ctx, &entry)
	if err != nil {
		return err
	}

	return nil
}
