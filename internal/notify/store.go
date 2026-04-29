package notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"
)

type Recipient struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	URL  string `json:"url"`
	Type string `json:"type"` // "slack", "teams", "webhook"
}

func (r Recipient) TypeLabel() string {
	switch r.Type {
	case "slack":
		return "Slack"
	case "teams":
		return "Teams"
	}
	return "Webhook"
}

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

func (r AlarmRule) CurrentValueStr() string {
	if r.LastValue == nil {
		return ""
	}
	return fmt.Sprintf("%.0f", *r.LastValue)
}

type VhostConfig struct {
	Recipients []Recipient `json:"recipients"`
	Rules      []AlarmRule `json:"rules"`
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

type Store struct {
	mu                  sync.RWMutex
	path                string
	Vhosts              map[string]*VhostConfig `json:"vhosts"`
	NotifiedMaintenance map[string]bool         `json:"notified_maintenance"`
}

func NewStore(path string) (*Store, error) {
	s := &Store{
		path:                path,
		Vhosts:              make(map[string]*VhostConfig),
		NotifiedMaintenance: make(map[string]bool),
	}
	if err := s.load(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	return s, nil
}

func (s *Store) load() error {
	data, err := os.ReadFile(s.path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, s)
}

func (s *Store) save() error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0644)
}

func (s *Store) GetVhostCopy(vhost string) VhostConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if vc, ok := s.Vhosts[vhost]; ok {
		recs := make([]Recipient, len(vc.Recipients))
		copy(recs, vc.Recipients)
		rules := make([]AlarmRule, len(vc.Rules))
		copy(rules, vc.Rules)
		return VhostConfig{Recipients: recs, Rules: rules}
	}
	return VhostConfig{}
}

func (s *Store) AllSnapshots() map[string]VhostConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make(map[string]VhostConfig, len(s.Vhosts))
	for vhost, vc := range s.Vhosts {
		recs := make([]Recipient, len(vc.Recipients))
		copy(recs, vc.Recipients)
		rules := make([]AlarmRule, len(vc.Rules))
		copy(rules, vc.Rules)
		out[vhost] = VhostConfig{Recipients: recs, Rules: rules}
	}
	return out
}

func (s *Store) AddRecipient(vhost, name, url, recipientType string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.Vhosts[vhost] == nil {
		s.Vhosts[vhost] = &VhostConfig{}
	}
	s.Vhosts[vhost].Recipients = append(s.Vhosts[vhost].Recipients, Recipient{
		ID:   fmt.Sprintf("%d", time.Now().UnixNano()),
		Name: name,
		URL:  url,
		Type: recipientType,
	})
	return s.save()
}

func (s *Store) DeleteRecipient(vhost, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if vc, ok := s.Vhosts[vhost]; ok {
		for i, r := range vc.Recipients {
			if r.ID == id {
				vc.Recipients = append(vc.Recipients[:i], vc.Recipients[i+1:]...)
				break
			}
		}
	}
	return s.save()
}

func (s *Store) AddRule(vhost string, rule AlarmRule) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.Vhosts[vhost] == nil {
		s.Vhosts[vhost] = &VhostConfig{}
	}
	rule.ID = fmt.Sprintf("%d", time.Now().UnixNano())
	rule.Status = "unknown"
	rule.Enabled = true
	s.Vhosts[vhost].Rules = append(s.Vhosts[vhost].Rules, rule)
	return s.save()
}

func (s *Store) DeleteRule(vhost, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if vc, ok := s.Vhosts[vhost]; ok {
		for i, r := range vc.Rules {
			if r.ID == id {
				vc.Rules = append(vc.Rules[:i], vc.Rules[i+1:]...)
				break
			}
		}
	}
	return s.save()
}

func (s *Store) ToggleRule(vhost, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if vc, ok := s.Vhosts[vhost]; ok {
		for i, r := range vc.Rules {
			if r.ID == id {
				vc.Rules[i].Enabled = !r.Enabled
				break
			}
		}
	}
	return s.save()
}

func (s *Store) GetRuleCopy(vhost, id string) (AlarmRule, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if vc, ok := s.Vhosts[vhost]; ok {
		for _, r := range vc.Rules {
			if r.ID == id {
				return r, true
			}
		}
	}
	return AlarmRule{}, false
}

func (s *Store) UpdateMessage(vhost, id, message string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if vc, ok := s.Vhosts[vhost]; ok {
		for i, r := range vc.Rules {
			if r.ID == id {
				vc.Rules[i].Message = message
				break
			}
		}
	}
	return s.save()
}

func (s *Store) UpdateRule(vhost, id, message string, threshold float64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if vc, ok := s.Vhosts[vhost]; ok {
		for i, r := range vc.Rules {
			if r.ID == id {
				vc.Rules[i].Message = message
				vc.Rules[i].Threshold = threshold
				break
			}
		}
	}
	return s.save()
}

func (s *Store) SetRuleStatus(vhost, id, status string, updateFiredTime bool, value *float64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if vc, ok := s.Vhosts[vhost]; ok {
		for i, r := range vc.Rules {
			if r.ID == id {
				vc.Rules[i].Status = status
				vc.Rules[i].LastValue = value
				if updateFiredTime {
					now := time.Now()
					vc.Rules[i].LastFired = &now
				}
				break
			}
		}
	}
	return s.save()
}

func (s *Store) IsMaintenanceNotified(id string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.NotifiedMaintenance[id]
}

func (s *Store) MarkMaintenanceNotified(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.NotifiedMaintenance == nil {
		s.NotifiedMaintenance = make(map[string]bool)
	}
	s.NotifiedMaintenance[id] = true
	return s.save()
}

// SendWebhooks posts a JSON message to each webhook URL.
// Uses {"text": "..."} format, compatible with Slack, Teams, and most webhook services.
func (s *Store) SendWebhooks(urls []string, subject, body string) error {
	text := subject + "\n\n" + body
	payload, _ := json.Marshal(map[string]string{"text": text})
	var lastErr error
	for _, u := range urls {
		resp, err := http.Post(u, "application/json", bytes.NewReader(payload))
		if err != nil {
			lastErr = err
			continue
		}
		resp.Body.Close()
		if resp.StatusCode >= 400 {
			lastErr = fmt.Errorf("webhook returnerte HTTP %d", resp.StatusCode)
		}
	}
	return lastErr
}

func BuildMessage(rule AlarmRule, vhost string) string {
	if rule.Message != "" {
		return rule.Message
	}
	switch rule.Type {
	case "channels":
		return fmt.Sprintf("Alarm utløst for vhost «%s».\n\nAntall channels har nådd grensen på %.0f.", vhost, rule.Threshold)
	case "connections":
		return fmt.Sprintf("Alarm utløst for vhost «%s».\n\nAntall connections har nådd grensen på %.0f.", vhost, rule.Threshold)
	case "queues":
		return fmt.Sprintf("Alarm utløst for vhost «%s».\n\nAntall køer har nådd grensen på %.0f.", vhost, rule.Threshold)
	case "unacked":
		return fmt.Sprintf("Alarm utløst for vhost «%s».\n\nAntall unacked meldinger har nådd grensen på %.0f.", vhost, rule.Threshold)
	case "queue_messages":
		return fmt.Sprintf("Alarm utløst for vhost «%s», kø «%s».\n\nAntall meldinger har nådd grensen på %.0f.", vhost, rule.QueueName, rule.Threshold)
	case "queue_size":
		return fmt.Sprintf("Alarm utløst for vhost «%s», kø «%s».\n\nKø-størrelse har nådd grensen på %.0f bytes.", vhost, rule.QueueName, rule.Threshold)
	case "no_consumer":
		return fmt.Sprintf("Alarm utløst for vhost «%s», kø «%s».\n\nDet er meldinger i køen, men ingen aktive consumers.", vhost, rule.QueueName)
	case "maintenance":
		return fmt.Sprintf("Alarm utløst for vhost «%s».\n\nDet er lagt ut en ny vedlikeholdsmelding.", vhost)
	}
	return fmt.Sprintf("Alarm «%s» ble utløst for vhost «%s».", rule.Name, vhost)
}
