package models

import (
	"fmt"
	"time"
)

type Recipient struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	URL  string `json:"url"`
	Type string `json:"type"` // "slack", "teams", "webhook"
}

func (r *Recipient) TypeLabel() string {
	switch r.Type {
	case "slack":
		return "Slack"
	case "teams":
		return "Teams"
	}
	return "Webhook"
}

// TODO: type to RuleType
type RuleType string

const ()

type AlarmRule struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	Type      string     `json:"type"`
	QueueName string     `json:"queue_name,omitempty"`
	Threshold float64    `json:"threshold,omitempty"`
	Message   string     `json:"message"`
	Enabled   bool       `json:"enabled"`
	Status    string     `json:"status"`
	LastFired *time.Time `json:"last_fired,omitempty"`
	LastValue *float64   `json:"last_value,omitempty"`
}

func (r AlarmRule) TypeLabel() string {
	switch r.Type {
	case "channels":
		return "Channels"
	case "connections":
		return "Connections"
	case "queues":
		return "Køer"
	case "unacked":
		return "Unacked meldinger"
	case "queue_messages":
		return "Meldinger i kø"
	case "queue_size":
		return "Kø-størrelse"
	case "no_consumer":
		return "Ingen consumer"
	case "maintenance":
		return "Vedlikeholdsmelding"
	}
	return r.Type
}

func (r AlarmRule) HasQueue() bool {
	return r.Type == "queue_messages" || r.Type == "queue_size" || r.Type == "no_consumer"
}

func (r AlarmRule) HasThreshold() bool {
	return r.Type != "no_consumer" && r.Type != "maintenance"
}

func (r AlarmRule) LastFiredStr() string {
	if r.LastFired == nil {
		return "—"
	}
	return r.LastFired.Format("2006-01-02 15:04")
}

func (r *AlarmRule) CurrentValueStr() string {
	if r.LastValue == nil {
		return ""
	}
	return fmt.Sprintf("%.0f", *r.LastValue)
}

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
