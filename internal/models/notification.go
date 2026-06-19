package models

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// @ ID unique identifier for the recipient
// @ Name human-readable name for the recipient
// @ URL webhook URL for the recipient - Slack, Teams
// @ Type type of the recipient - "slack", "teams", "webhook"
type Recipient struct {
	ID   string        `json:"id"`
	Name string        `json:"name"`
	URL  string        `json:"url"`
	Type RecipientType `json:"type"`
}

func (r *Recipient) UnmarshalJSON(data []byte) error {
	var aux struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		URL  string `json:"url"`
		Type string `json:"type"`
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	r.ID = aux.ID
	r.Name = aux.Name
	r.URL = aux.URL
	r.Type = ParseRecipientType(aux.Type)
	if r.Type == RecipientTypeUnknown {
		return fmt.Errorf("invalid recipient type: %s, expected one of %s", aux.Type, GetRecipientTypesString())
	}

	return nil
}

type RecipientType string

const (
	RecipientTypeSlack   RecipientType = "slack"   //	@name	Slack
	RecipientTypeTeams   RecipientType = "teams"   //	@name	Teams
	RecipientTypeWebhook RecipientType = "webhook" //	@name	Webhook
	RecipientTypeUnknown RecipientType = "unknown"
)

func GetReceipientTypes() []RecipientType {
	return []RecipientType{
		RecipientTypeSlack,
		RecipientTypeTeams,
		RecipientTypeWebhook,
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
	case "slack":
		return RecipientTypeSlack
	case "teams":
		return RecipientTypeTeams
	case "webhook":
		return RecipientTypeWebhook
	default:
		return RecipientTypeUnknown
	}
}

// Alarm rule definition
// @ Description ID unique identifier for the alarm rule
// @ Description Name human-readable name for the alarm rule
// @ Description Type type of the alarm (e.g., "channels", "connections", "queues", etc.)
// @ Description QueueName optional name of the queue - required for "queues" alarm types
// @ Description Threshold numeric threshold that triggers the alarm
// @ Description Message custom message to include in the notification when the alarm is triggered
// @ Description Enabled indicates whether the alarm rule is active
type AlarmRule struct {
	ID        string      `json:"id" bson:"_id"`
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

type AlarmRuleUpdate struct {
	Threshold float64 `json:"threshold" bson:"threshold"`
	Message   string  `json:"message" bson:"message"`
}

func (a *AlarmRule) UnmarshalJSON(data []byte) error {
	var Alias struct {
		ID        string      `json:"id" bson:"_id"`
		Name      string      `json:"name" bson:"name"`
		Type      AlarmType   `json:"type" bson:"type"`
		QueueName string      `json:"queue_name" bson:"queueName"`
		Threshold float64     `json:"threshold" bson:"threshold"`
		Message   string      `json:"message" bson:"message"`
		Enabled   bool        `json:"enabled" bson:"enabled"`
		Status    AlarmStatus `json:"status" bson:"status"`
		LastFired string      `json:"last_fired" bson:"lastFired"`
		LastValue *float64    `json:"last_value" bson:"lastValue"`
	}
	if err := json.Unmarshal(data, &Alias); err != nil {
		return err
	}

	a.ID = Alias.ID
	a.Name = Alias.Name
	if !isValidAlarmType(string(Alias.Type)) {
		return fmt.Errorf("invalid alarm type: %s", Alias.Type)
	}
	a.Type = Alias.Type
	a.QueueName = Alias.QueueName
	a.Threshold = Alias.Threshold
	a.Message = Alias.Message
	a.Enabled = Alias.Enabled
	a.Status = Alias.Status
	time, err := time.Parse(time.RFC3339, Alias.LastFired)
	if err != nil {
		return fmt.Errorf("invalid time format for last_fired: %w", err)
	}
	a.LastFired = &time
	a.LastValue = Alias.LastValue

	return nil
}

type AlarmStatus string

const (
	AlarmStatusActive   AlarmStatus = "active"   // @ name Active
	AlarmStatusInactive AlarmStatus = "inactive" // @ name Inactive
	AlarmStatusFired    AlarmStatus = "fired"    // @ name Fired
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

func isValidAlarmType(s string) bool {
	for _, t := range GetAlarmTypes() {
		if string(t) == s {
			return true
		}
	}
	return false
}

// TODO: Rewrite to english and if possible make more generic
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

type (
	VhostConfig struct {
		Recipients []Recipient `json:"recipients"`
		Rules      []AlarmRule `json:"rules"`
	}
	VhostConfigOptions func(*VhostConfig)
)

func WithRecipients(recipients []Recipient) VhostConfigOptions {
	return func(vc *VhostConfig) {
		vc.Recipients = recipients
	}
}

func WithRules(rules []AlarmRule) VhostConfigOptions {
	return func(vc *VhostConfig) {
		vc.Rules = rules
	}
}

func NewVhostConfig(opts ...VhostConfigOptions) *VhostConfig {
	vc := &VhostConfig{}
	for _, opt := range opts {
		opt(vc)
	}
	return vc
}

func (vc VhostConfig) WebhookURLs() []string {
	var urls []string
	for _, r := range vc.Recipients {
		if r.URL != "" {
			urls = append(urls, r.URL)
		}
	}
	return urls
}

type TestNotificationResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type VhostNotification struct {
	Name       string      `bson:"_id"`
	Recipients []Recipient `bson:"recipients"`
	Rules      []AlarmRule `bson:"rules"`
	Notified   bool        `bson:"notified"`
}

func NewVhostNotification(name string) *VhostNotification {
	return &VhostNotification{
		Name:       name,
		Recipients: make([]Recipient, 0),
		Rules:      make([]AlarmRule, 0),
		Notified:   false,
	}
}

func (vn *VhostNotification) WebhookURLs() []string {
	urls := make([]string, 0, len(vn.Recipients))
	for _, r := range vn.Recipients {
		if r.Type == RecipientTypeWebhook {
			urls = append(urls, r.URL)
		}
	}
	return urls
}

type NotificationResponse struct {
	Vhost         *Vhost             `json:"vhost"`
	Notifications *VhostNotification `json:"notifications"`
}

func NewNotificationResponse(vhost *Vhost, notifications *VhostNotification) *NotificationResponse {
	return &NotificationResponse{
		Vhost:         vhost,
		Notifications: notifications,
	}
}
