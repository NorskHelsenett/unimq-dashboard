package models

import (
	"fmt"
	"strings"
	"time"
)

type Recipient struct {
	ID   string        `json:"id"`
	Name string        `json:"name"`
	URL  string        `json:"url"`
	Type RecipientType `json:"type"` // "slack", "teams", "webhook"
}

type RecipientType string

const (
	RecipientTypeSlack   RecipientType = "slack"
	RecipientTypeTeams   RecipientType = "teams"
	RecipientTypeWebhook RecipientType = "webhook"
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

type AlarmRule struct {
	ID        string     `json:"id" bson:"_id"`
	Name      string     `json:"name" bson:"name"`
	Type      AlarmType  `json:"type" bson:"type"`
	QueueName string     `json:"queue_name,omitempty" bson:"queueName"`
	Threshold float64    `json:"threshold,omitempty" bson:"threshold"`
	Message   string     `json:"message" bson:"message"`
	Enabled   bool       `json:"enabled" bson:"enabled"`
	Status    string     `json:"status" bson:"status"` // TODO: Enum "active", "inactive", "fired"
	LastFired *time.Time `json:"last_fired,omitempty" bson:"lastFired"`
	LastValue *float64   `json:"last_value,omitempty" bson:"lastValue"`
}

type AlarmType string

const (
	AlarmTypeChannels      AlarmType = "channels"
	AlarmTypeConnections   AlarmType = "connections"
	AlarmTypeQueues        AlarmType = "queues"
	AlarmTypeUnacked       AlarmType = "unacked"
	AlarmTypeQueueMessages AlarmType = "queue_messages"
	AlarmTypeQueueSize     AlarmType = "queue_size"
	AlarmTypeNoConsumer    AlarmType = "no_consumer"
	AlarmTypeMaintenance   AlarmType = "maintenance"
)

// func (r AlarmRule) TypeLabel() string {
// 	switch r.Type {
// 	case "channels":
// 		return "Channels"
// 	case "connections":
// 		return "Connections"
// 	case "queues":
// 		return "Køer"
// 	case "unacked":
// 		return "Unacked meldinger"
// 	case "queue_messages":
// 		return "Meldinger i kø"
// 	case "queue_size":
// 		return "Kø-størrelse"
// 	case "no_consumer":
// 		return "Ingen consumer"
// 	case "maintenance":
// 		return "Vedlikeholdsmelding"
// 	}
// 	return r.Type
// }

// func (r AlarmRule) HasQueue() bool {
// 	return r.Type == "queue_messages" || r.Type == "queue_size" || r.Type == "no_consumer"
// }
//
// func (r AlarmRule) HasThreshold() bool {
// 	return r.Type != "no_consumer" && r.Type != "maintenance"
// }
//
// func (r AlarmRule) LastFiredStr() string {
// 	if r.LastFired == nil {
// 		return "—"
// 	}
// 	return r.LastFired.Format("2006-01-02 15:04")
// }
//
// func (r *AlarmRule) CurrentValueStr() string {
// 	if r.LastValue == nil {
// 		return ""
// 	}
// 	return fmt.Sprintf("%.0f", *r.LastValue)
// }

func (r *AlarmRule) BuildMessage(vhost string) string {
	if r.Message != "" {
		return r.Message
	}
	switch r.Type {
	case "channels":
		return fmt.Sprintf("Alarm utløst for vhost «%s».\n\nAntall channels har nådd grensen på %.0f.", vhost, r.Threshold)
	case "connections":
		return fmt.Sprintf("Alarm utløst for vhost «%s».\n\nAntall connections har nådd grensen på %.0f.", vhost, r.Threshold)
	case "queues":
		return fmt.Sprintf("Alarm utløst for vhost «%s».\n\nAntall køer har nådd grensen på %.0f.", vhost, r.Threshold)
	case "unacked":
		return fmt.Sprintf("Alarm utløst for vhost «%s».\n\nAntall unacked meldinger har nådd grensen på %.0f.", vhost, r.Threshold)
	case "queue_messages":
		return fmt.Sprintf("Alarm utløst for vhost «%s», kø «%s».\n\nAntall meldinger har nådd grensen på %.0f.", vhost, r.QueueName, r.Threshold)
	case "queue_size":
		return fmt.Sprintf("Alarm utløst for vhost «%s», kø «%s».\n\nKø-størrelse har nådd grensen på %.0f bytes.", vhost, r.QueueName, r.Threshold)
	case "no_consumer":
		return fmt.Sprintf("Alarm utløst for vhost «%s», kø «%s».\n\nDet er meldinger i køen, men ingen aktive consumers.", vhost, r.QueueName)
	case "maintenance":
		return fmt.Sprintf("Alarm utløst for vhost «%s».\n\nDet er lagt ut en ny vedlikeholdsmelding.", vhost)
	}
	return fmt.Sprintf("Alarm «%s» ble utløst for vhost «%s».", r.Name, vhost)
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
