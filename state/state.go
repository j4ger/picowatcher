package state

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

// State maps feed URL -> set of seen item IDs
type State struct {
	mu   sync.Mutex
	path string
	Seen map[string]map[string]bool `json:"seen"`
}

// Load reads persisted state from path. The returned *State is not safe for
// concurrent use until Load returns; callers must not share it across goroutines
// before that point.
func Load(path string) (*State, error) {
	s := &State{
		path: path,
		Seen: make(map[string]map[string]bool),
	}
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return s, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading state file: %w", err)
	}
	if err := json.Unmarshal(data, &s.Seen); err != nil {
		return nil, fmt.Errorf("parsing state file: %w", err)
	}
	return s, nil
}

func (s *State) Save() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	data, err := json.MarshalIndent(s.Seen, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling state: %w", err)
	}
	if err := os.WriteFile(s.path, data, 0644); err != nil {
		return fmt.Errorf("writing state file: %w", err)
	}
	return nil
}

func (s *State) IsNew(feedURL, itemID string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	feedSeen, ok := s.Seen[feedURL]
	if !ok {
		return true
	}
	return !feedSeen[itemID]
}

func (s *State) MarkSeen(feedURL, itemID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.Seen[feedURL]; !ok {
		s.Seen[feedURL] = make(map[string]bool)
	}
	s.Seen[feedURL][itemID] = true
}
