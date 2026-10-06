package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sisneve/rabbitmq-dashboard/internal/api/httpsuite"
)

var (
	ErrNoFieldsToUpdate = fmt.Errorf("no fields to update")
)

// Alarm rule definition
type AlarmRulePatch struct {
	Name      httpsuite.Optional[string]    `json:"name" swaggertype:"string" example:"High Queue Size" validate:"nonull"`
	Type      httpsuite.Optional[AlarmType] `json:"type" swaggertype:"string"  example:"queue_size" validate:"alarmtype,nonull"`
	QueueName httpsuite.Optional[string]    `json:"queue_name,omitempty" eswaggertype:"string" xample:"my-queue" validate:"omitempty,nonull"`
	Threshold httpsuite.Optional[float64]   `json:"threshold,omitempty" swaggertype:"number" example:"1000" validate:"nonull"`
	Message   httpsuite.Optional[string]    `json:"message" swaggertype:"string" example:"Queue size has exceeded the threshold"`
	Enabled   httpsuite.Optional[bool]      `json:"enabled" swaggertype:"boolean" example:"true" validate:"nonull"`
}

func ParseAlarmRulePatch(patch *AlarmRulePatch) (map[string]any, error) {
	setFields := map[string]any{}
	httpsuite.SetUpdate(setFields, "rules.$.name", patch.Name)
	httpsuite.SetUpdate(setFields, "rules.$.type", patch.Type)
	httpsuite.SetUpdate(setFields, "rules.$.queueName", patch.QueueName)
	httpsuite.SetUpdate(setFields, "rules.$.threshold", patch.Threshold)
	httpsuite.SetUpdate(setFields, "rules.$.message", patch.Message)
	httpsuite.SetUpdate(setFields, "rules.$.enabled", patch.Enabled)

	if len(setFields) == 0 {
		return nil, ErrNoFieldsToUpdate
	}

	return setFields, nil
}

type AlarmRuleCreate struct {
	Name      string    `json:"name" bson:"name" example:"High Queue Size"`
	Type      AlarmType `json:"type" bson:"type" example:"queue_size"`
	QueueName string    `json:"queue_name,omitempty" bson:"queueName" example:"my-queue"`
	Threshold float64   `json:"threshold,omitempty" bson:"threshold" example:"1000"`
	Message   string    `json:"message" bson:"message" example:"Queue size has exceeded the threshold" validate:"nonull"`
	Enabled   bool      `json:"enabled" bson:"enabled" example:"true"`
}

var (
	ErrInvalidAlarmType = fmt.Errorf("invalid alarm type")
)

func (c *AlarmRuleCreate) ToAlarmRule() (*AlarmRule, error) {
	if !IsValidAlarmType(string(c.Type)) {
		return nil, fmt.Errorf("%w: %s, expected one of %s", ErrInvalidAlarmType, c.Type, GetAlarmTypes())
	}

	var status AlarmStatus
	if c.Enabled {
		status = AlarmStatusActive
	} else {
		status = AlarmStatusInactive
	}
	return &AlarmRule{
		ID:        uuid.New().String(),
		Name:      c.Name,
		Type:      c.Type,
		QueueName: c.QueueName,
		Threshold: c.Threshold,
		Message:   c.Message,
		Enabled:   c.Enabled,
		Status:    status,
		LastFired: nil,
		LastValue: nil,
	}, nil
}

type AlarmRule struct {
	ID        string      `json:"id" bson:"id"`
	Name      string      `json:"name" bson:"name"`
	Type      AlarmType   `json:"type" bson:"type"`
	QueueName string      `json:"queue_name,omitempty" bson:"queueName"`
	Threshold float64     `json:"threshold,omitempty" bson:"threshold"`
	Message   string      `json:"message" bson:"message"`
	Enabled   bool        `json:"enabled" bson:"enabled"`
	Status    AlarmStatus `json:"status" bson:"status"`
	LastFired *time.Time  `json:"last_fired,omitempty" bson:"lastFired"`
	LastValue *float64    `json:"last_value,omitempty" bson:"lastValue"`
}

func NewAlarmRule(name string, typ AlarmType, queueName string, threshold float64, message string, enabled bool) *AlarmRule {
	return &AlarmRule{
		ID:        uuid.New().String(),
		Name:      name,
		Type:      typ,
		QueueName: queueName,
		Threshold: threshold,
		Message:   message,
		Enabled:   enabled,
		Status:    AlarmStatusActive,
		LastFired: nil,
		LastValue: nil,
	}
}

func (r *AlarmRule) IsTriggered(value float64) bool {
	return value >= r.Threshold
}

func (r *AlarmRule) IsChanged(other *AlarmRule) bool {
	if r.Name != other.Name {
		return true
	}
	if r.Type != other.Type {
		return true
	}
	if r.QueueName != other.QueueName {
		return true
	}
	if r.Threshold != other.Threshold {
		return true
	}
	if r.Message != other.Message {
		return true
	}
	if r.Enabled != other.Enabled {
		return true
	}
	return false
}

func (r *AlarmRule) BuildMessage(vhost string) string {
	if r.Message != "" {
		return r.Message
	}
	base := fmt.Sprintf("Alarm «%s» triggered for vhost '%s'", r.Name, vhost)
	trigger := fmt.Sprintf("Number of %v has reached the threshold of %.0f.", r.Type, r.Threshold)
	switch r.Type {
	case "channels":
		return fmt.Sprintf("%v.\n\n%v", base, trigger)
	case "connections":
		return fmt.Sprintf("%v.\n\n%v", base, trigger)
	case "queues":
		return fmt.Sprintf("%v.\n\n%v", base, trigger)
	case "unacked":
		return fmt.Sprintf("%v.\n\n%v", base, trigger)
	case "queue_messages":
		return fmt.Sprintf("%v, queue '%s'.\n\n%v", base, r.QueueName, trigger)
	case "queue_size":
		return fmt.Sprintf("%v, queue '%s'.\n\n%v", base, r.QueueName, trigger)
	case "no_consumer":
		return fmt.Sprintf("%v, queue '%s'.\n\nThere's messages in the queue, but no consumers.", base, r.QueueName)
	case "maintenance":
		return fmt.Sprintf("%v.\n\nA new maintenance window has been scheduled.", base)
	}
	return fmt.Sprintf("Alarm '%s' has been triggered for vhost '%s'.", r.Name, vhost)
}
