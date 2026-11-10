package models

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/sisneve/rabbitmq-dashboard/internal/api/httpsuite"
)

// PostRecipient is used for creating a new recipient
//
//		@Name	human-readable name for the recipient
//		@URL	webhook URL for the recipient - Slack, Teams
//		@Email	email address for the recipient - used for email notifications
//		@Type	type of the recipient - "webhook", "email"
//	 @Notified	whether the recipient has been notified or not
type PostRecipient struct {
	Name     string        `json:"name" bson:"name" example:"Slack Channel to team" validate:"required"`
	URL      string        `json:"url" bson:"url" example:"https://hooks.slack.com/services" validate:"omitempty,url"`
	Email    string        `json:"email" bson:"email" example:"ola.normann@normann.no" validate:"omitempty,email"`
	Type     RecipientType `json:"type" bson:"type" example:"webhook" validate:"required,oneof=webhook email"`
	Notified bool          `json:"notified" bson:"notified" example:"false" validate:"required"`
}

func (p *PostRecipient) ToRecipient() (*Recipient, error) {

	typ := ParseRecipientType(string(p.Type))
	if typ == RecipientTypeUnknown {
		return nil, fmt.Errorf("invalid recipient type: %s, expected one of %s", p.Type, GetRecipientTypesString())
	}
	return &Recipient{
		ID:       uuid.New().String(),
		Name:     p.Name,
		URL:      p.URL,
		Email:    p.Email,
		Type:     typ,
		Notified: p.Notified,
	}, nil
}

// @ID		unique identifier for the recipient
// @Name	human-readable name for the recipient
// @URL	webhook URL for the recipient - Slack, Teams
// @Email	email address for the recipient - used for email notifications
// @Type	type of the recipient - "webhook", "webhook"
// @Notified	whether the recipient has been notified or not
type Recipient struct {
	ID       string        `json:"id" validate:"required" bson:"_id"`
	Vhost    string        `json:"vhost" validate:"required" bson:"vhost"`
	Name     string        `json:"name" validate:"required" bson:"name"`
	URL      string        `json:"url,omitempty" validate:"omitempty,url" bson:"url,omitempty"`
	Email    string        `json:"email,omitempty" validate:"omitempty,email" bson:"email,omitempty"`
	Type     RecipientType `json:"type" validate:"required,oneof=webhook email" bson:"type"`
	Notified bool          `json:"notified" validate:"required" bson:"notified"`
}

type RecipientPost struct {
	Name     httpsuite.Optional[string]        `json:"name" validate:"required"`
	URL      httpsuite.Optional[string]        `json:"url,omitempty" validate:"omitempty,url"`
	Email    httpsuite.Optional[string]        `json:"email,omitempty" validate:"omitempty,email"`
	Type     httpsuite.Optional[RecipientType] `json:"type" validate:"required,oneof=webhook email"`
	Notified httpsuite.Optional[bool]          `json:"notified" validate:"required" validate:"nonull"`
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
