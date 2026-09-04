package models

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// PostRecipient is used for creating a new recipient
//
//	@Name	human-readable name for the recipient
//	@URL	webhook URL for the recipient - Slack, Teams
//	@Email	email address for the recipient - used for email notifications
//	@Type	type of the recipient - "webhook", "email"
type PostRecipient struct {
	Name  string        `json:"name" bson:"name" example:"Slack Channel to team"`
	URL   string        `json:"url" bson:"url" example:"https://hooks.slack.com/services"`
	Email string        `json:"email" bson:"email" example:"ola.normann@normann.no"`
	Type  RecipientType `json:"type" bson:"type" example:"webhook"`
}

func (p *PostRecipient) ToRecipient() (*Recipient, error) {

	typ := ParseRecipientType(string(p.Type))
	if typ == RecipientTypeUnknown {
		return nil, fmt.Errorf("invalid recipient type: %s, expected one of %s", p.Type, GetRecipientTypesString())
	}
	return &Recipient{
		ID:    uuid.New().String(),
		Name:  p.Name,
		URL:   p.URL,
		Email: p.Email,
		Type:  typ,
	}, nil
}

// @ID		unique identifier for the recipient
// @Name	human-readable name for the recipient
// @URL	webhook URL for the recipient - Slack, Teams
// @Type	type of the recipient - "slack", "teams", "webhook"
type Recipient struct {
	ID    string        `json:"id"`
	Name  string        `json:"name"`
	URL   string        `json:"url,omitempty"`
	Email string        `json:"email,omitempty"`
	Type  RecipientType `json:"type"`
}

type RecipientType string

const (
	RecipientTypeWebhook RecipientType = "webhook" //	@name	Webhook
	RecipientTypeEmail   RecipientType = "email"   //	@name	Email
	RecipientTypeUnknown RecipientType = "unknown"
)

func GetReceipientTypes() []RecipientType {
	return []RecipientType{
		RecipientTypeWebhook,
		RecipientTypeEmail,
	}
}

func GetRecipientTypesString() string {
	types := GetReceipientTypes()
	strs := make([]string, len(types))
	for i, t := range types {
		strs[i] = string(t)
	}
	return fmt.Sprintf("[%s]", strings.Join(strs, ", "))
}

func ParseRecipientType(s string) RecipientType {
	switch s {
	case "webhook":
		return RecipientTypeWebhook
	case "email":
		return RecipientTypeEmail
	default:
		return RecipientTypeUnknown
	}
}

// Alarm rule definition
type PostAlarmRule struct {
	Name      string    `json:"name" bson:"name" example:"High Queue Size"`
	Type      AlarmType `json:"type" bson:"type" example:"queue_size"`
	QueueName string    `json:"queue_name,omitempty" bson:"queueName" example:"my-queue"`
	Threshold float64   `json:"threshold,omitempty" bson:"threshold" example:"1000"`
	Message   string    `json:"message" bson:"message" example:"Queue size has exceeded the threshold"`
	Enabled   bool      `json:"enabled" bson:"enabled" example:"true"`
}

func (p *PostAlarmRule) ToAlarmRule() (*AlarmRule, error) {
	if !IsValidAlarmType(string(p.Type)) {
		return nil, fmt.Errorf("invalid alarm type: %s", p.Type)
	}
	return &AlarmRule{
		ID:        uuid.New().String(),
		Name:      p.Name,
		Type:      p.Type,
		QueueName: p.QueueName,
		Threshold: p.Threshold,
		Message:   p.Message,
		Enabled:   p.Enabled,
		Status:    AlarmStatusActive,
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

type AlarmRuleUpdate struct {
	Threshold float64 `json:"threshold" bson:"threshold" example:"1000"`
	Message   string  `json:"message" bson:"message" example:"Queue size has exceeded the threshold"`
}

type AlarmStatus string

const (
	AlarmStatusOK       AlarmStatus = "ok"       //	@name	OK
	AlarmStatusActive   AlarmStatus = "active"   //	@name	Active
	AlarmStatusInactive AlarmStatus = "inactive" //	@name	Inactive
	AlarmStatusFiring   AlarmStatus = "firing"   //	@name	Firing
	AlarmStatusFired    AlarmStatus = "fired"    //	@name	Fired
	AlarmStatusUnknown  AlarmStatus = "unknown"  //
)

type AlarmType string

const (
	AlarmTypeChannels      AlarmType = "channels"       //	@name	Channels
	AlarmTypeConnections   AlarmType = "connections"    //	@name	Connections
	AlarmTypeQueues        AlarmType = "queues"         //	@name	Queues
	AlarmTypeUnacked       AlarmType = "unacked"        //	@name	Unacked_Messages
	AlarmTypeQueueMessages AlarmType = "queue_messages" //	@name	Queue_Messages
	AlarmTypeQueueSize     AlarmType = "queue_size"     //	@name	Queue_Size
	AlarmTypeNoConsumer    AlarmType = "no_consumer"    //	@name	No_Consumer
	AlarmTypeMaintenance   AlarmType = "maintenance"    //	@name	Maintenance
)

func GetAlarmTypes() []AlarmType {
	return []AlarmType{
		AlarmTypeChannels,
		AlarmTypeConnections,
		AlarmTypeQueues,
		AlarmTypeUnacked,
		AlarmTypeQueueMessages,
		AlarmTypeQueueSize,
		AlarmTypeNoConsumer,
		AlarmTypeMaintenance,
	}
}

func IsValidAlarmType(s string) bool {
	for _, t := range GetAlarmTypes() {
		if string(t) == s {
			return true
		}
	}
	return false
}

func ConvertToAlarmType(s string) (AlarmType, error) {
	for _, t := range GetAlarmTypes() {
		if string(t) == s {
			return t, nil
		}
	}
	return "", fmt.Errorf("invalid alarm type: %s", s)
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

type TestNotificationResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type VhostNotification struct {
	Name       string       `bson:"_id"`
	Recipients []*Recipient `bson:"recipients"`
	Rules      []*AlarmRule `bson:"rules"`
	Notified   bool         `bson:"notified"`
}

func NewVhostNotification(name string) *VhostNotification {
	return &VhostNotification{
		Name:       name,
		Recipients: make([]*Recipient, 0),
		Rules:      make([]*AlarmRule, 0),
		Notified:   false,
	}
}

func (vn *VhostNotification) WebhookURLs() []string {
	urls := make([]string, 0)
	for _, r := range vn.Recipients {
		if r.Type == RecipientTypeWebhook {
			if r.URL != "" {
				urls = append(urls, r.URL)
			}
		}
	}
	return urls
}

func (vn *VhostNotification) EmailRecipients() []string {
	emails := make([]string, 0)
	for _, r := range vn.Recipients {
		if r.Type == RecipientTypeEmail {
			if r.Email != "" {
				emails = append(emails, r.Email)
			}
		}
	}
	return emails
}
