package notificationhelper

type NotifyStatus struct {
	WebhookURLs     []string        `json:"webhook_urls"`
	WebhookStatuses []WebhookStatus `json:"webhook_statuses"`
	EmailRecipients []string        `json:"email_recipients"`
	EmailStatuses   []EmailStatus   `json:"email_statuses"`
}

func NewNotifyStatus(webhookURLs []string, emailRecipients []string) *NotifyStatus {
	return &NotifyStatus{
		WebhookURLs:     webhookURLs,
		WebhookStatuses: make([]WebhookStatus, len(webhookURLs)),
		EmailRecipients: emailRecipients,
		EmailStatuses:   make([]EmailStatus, len(emailRecipients)),
	}
}

func (nh *NotifyStatus) HasErrors() bool {
	for _, ws := range nh.WebhookStatuses {
		if !ws.OK {
			return true
		}
	}
	for _, es := range nh.EmailStatuses {
		if !es.OK {
			return true
		}
	}
	return false
}

func (nh *NotifyStatus) IsTotalFailure() bool {
	for _, ws := range nh.WebhookStatuses {
		if ws.OK {
			return false
		}
	}
	for _, es := range nh.EmailStatuses {
		if es.OK {
			return false
		}
	}
	return true
}

func (nh *NotifyStatus) IsPartialFailure() bool {
	return nh.HasErrors() && !nh.IsTotalFailure()
}

func (nh *NotifyStatus) IsTotalSuccess() bool {
	for _, ws := range nh.WebhookStatuses {
		if !ws.OK {
			return false
		}
	}
	for _, es := range nh.EmailStatuses {
		if !es.OK {
			return false
		}
	}
	return true
}

func (nh *NotifyStatus) FailedDestinations() []string {
	failed := make([]string, 0)
	for _, ws := range nh.WebhookStatuses {
		if !ws.OK {
			failed = append(failed, ws.URL)
		}
	}
	for _, es := range nh.EmailStatuses {
		if !es.OK {
			failed = append(failed, es.Recipient)
		}
	}
	return failed
}
