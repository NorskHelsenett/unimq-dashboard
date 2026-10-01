package models

import (
	"fmt"
)

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
	AlarmTypeChannels      AlarmType = "channels"       //	@name	channels
	AlarmTypeConnections   AlarmType = "connections"    //	@name	connections
	AlarmTypeQueues        AlarmType = "queues"         //	@name	queues
	AlarmTypeUnacked       AlarmType = "unacked"        //	@name	unacked_messages
	AlarmTypeQueueMessages AlarmType = "queue_messages" //	@name	queue_messages
	AlarmTypeQueueSize     AlarmType = "queue_size"     //	@name	queue_size
	AlarmTypeNoConsumer    AlarmType = "no_consumer"    //	@name	no_consumer
	AlarmTypeMaintenance   AlarmType = "maintenance"    //	@name	maintenance
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

func GetQueueAlarmTypes() []AlarmType {
	return []AlarmType{
		AlarmTypeQueueMessages,
		AlarmTypeQueueSize,
		AlarmTypeNoConsumer,
		AlarmTypeUnacked,
	}
}

func GetVhostAlarmTypes() []AlarmType {
	return []AlarmType{
		AlarmTypeChannels,
		AlarmTypeConnections,
		AlarmTypeQueues,
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

type TestNotificationResponse struct {
	Success            bool     `json:"success"`
	Message            string   `json:"message"`
	FailedDestinations []string `json:"failed_destinations,omitempty"`
}

type VhostNotification struct {
	Name       string       `bson:"_id" json:"name"`
	Recipients []*Recipient `bson:"recipients" json:"recipients"`
	Rules      []*AlarmRule `bson:"rules" json:"rules"`
	Notified   bool         `bson:"notified" json:"notified"`
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
