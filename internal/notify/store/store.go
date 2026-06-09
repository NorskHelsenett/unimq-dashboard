package store

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/sisneve/rabbitmq-dashboard/internal/models"
)

type Store struct {
	mu                  sync.RWMutex
	path                string
	Vhosts              map[string]*models.VhostConfig `json:"vhosts"`
	NotifiedMaintenance map[string]bool                `json:"notified_maintenance"`
}

func NewStore(path string) (*Store, error) {
	s := &Store{
		path:                path,
		Vhosts:              make(map[string]*models.VhostConfig),
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

func (s *Store) GetVhostCopy(vhost string) models.VhostConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if vc, ok := s.Vhosts[vhost]; ok {
		recs := make([]models.Recipient, len(vc.Recipients))
		copy(recs, vc.Recipients)
		rules := make([]models.AlarmRule, len(vc.Rules))
		copy(rules, vc.Rules)
		return *models.NewVhostConfig(models.WithRecipients(recs), models.WithRules(rules))
	}
	return *models.NewVhostConfig()
}

func (s *Store) AllSnapshots() map[string]models.VhostConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make(map[string]models.VhostConfig, len(s.Vhosts))
	for vhost, vc := range s.Vhosts {
		recs := make([]models.Recipient, len(vc.Recipients))
		copy(recs, vc.Recipients)
		rules := make([]models.AlarmRule, len(vc.Rules))
		copy(rules, vc.Rules)
		out[vhost] = *models.NewVhostConfig(models.WithRecipients(recs), models.WithRules(rules))
	}
	return out
}

func (s *Store) AddRecipient(vhost, name, url, recipientType string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.Vhosts[vhost] == nil {
		s.Vhosts[vhost] = models.NewVhostConfig()
	}
	s.Vhosts[vhost].Recipients = append(s.Vhosts[vhost].Recipients, models.Recipient{
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

func (s *Store) AddRule(vhost string, rule models.AlarmRule) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.Vhosts[vhost] == nil {
		s.Vhosts[vhost] = models.NewVhostConfig()
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

func (s *Store) GetRuleCopy(vhost, id string) (models.AlarmRule, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if vc, ok := s.Vhosts[vhost]; ok {
		for _, r := range vc.Rules {
			if r.ID == id {
				return r, true
			}
		}
	}
	return models.AlarmRule{}, false
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
