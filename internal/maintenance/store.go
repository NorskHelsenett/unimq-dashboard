package maintenance

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

type Entry struct {
	ID          string    `json:"id"`
	Description string    `json:"description"`
	Start       time.Time `json:"start"`
	End         time.Time `json:"end"`
	Status      string    `json:"status"` // "scheduled", "done", "skipped"
}

type Store struct {
	mu      sync.RWMutex
	path    string
	Entries []Entry `json:"entries"`
}

func NewStore(path string) (*Store, error) {
	s := &Store{path: path}
	if err := s.load(); err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
		// Seed with one historical entry
		s.Entries = []Entry{
			{
				ID:          "20260113120000",
				Description: "Oppgradering til RabbitMQ 4.2.2 og OS patch",
				Start:       time.Date(2026, 1, 13, 12, 0, 0, 0, time.UTC),
				End:         time.Date(2026, 1, 13, 13, 0, 0, 0, time.UTC),
				Status:      "done",
			},
		}
		if err := s.save(); err != nil {
			return nil, err
		}
	}
	return s, nil
}

func (s *Store) load() error {
	data, err := os.ReadFile(s.path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &s.Entries)
}

func (s *Store) save() error {
	data, err := json.MarshalIndent(s.Entries, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0644)
}

func (s *Store) Scheduled() []Entry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []Entry
	for _, e := range s.Entries {
		if e.Status == "scheduled" {
			out = append(out, e)
		}
	}
	return out
}

func (s *Store) History() []Entry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []Entry
	for _, e := range s.Entries {
		if e.Status != "scheduled" {
			out = append(out, e)
		}
	}
	return out
}

func (s *Store) All() []Entry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Entry, len(s.Entries))
	copy(out, s.Entries)
	return out
}

func (s *Store) Add(description, startStr, endStr string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	start, err := time.ParseInLocation("2006-01-02T15:04", startStr, time.Local)
	if err != nil {
		return fmt.Errorf("ugyldig starttidspunkt: %w", err)
	}
	end, err := time.ParseInLocation("2006-01-02T15:04", endStr, time.Local)
	if err != nil {
		return fmt.Errorf("ugyldig sluttidspunkt: %w", err)
	}

	e := Entry{
		ID:          time.Now().Format("20060102150405"),
		Description: description,
		Start:       start,
		End:         end,
		Status:      "scheduled",
	}
	s.Entries = append([]Entry{e}, s.Entries...)
	return s.save()
}

func (s *Store) SetStatus(id, status string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.Entries {
		if s.Entries[i].ID == id {
			s.Entries[i].Status = status
			return s.save()
		}
	}
	return nil
}

func (s *Store) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, e := range s.Entries {
		if e.ID == id {
			s.Entries = append(s.Entries[:i], s.Entries[i+1:]...)
			return s.save()
		}
	}
	return nil
}
