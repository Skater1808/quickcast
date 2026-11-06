package state

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Entry stores execution count and last access timestamp.
type Entry struct {
	Count      int   `json:"count"`
	LastAccess int64 `json:"last_access"` // Unix timestamp in seconds
}

// Store tracks usage statistics for frecency ranking.
type Store struct {
	mu       sync.RWMutex
	filePath string
	entries  map[string]Entry
}

// NewStore initializes or loads frecency state from disk.
func NewStore(customPath ...string) (*Store, error) {
	var targetPath string
	if len(customPath) > 0 && customPath[0] != "" {
		targetPath = customPath[0]
	} else {
		dataHome := os.Getenv("XDG_DATA_HOME")
		if dataHome == "" {
			home, _ := os.UserHomeDir()
			dataHome = filepath.Join(home, ".local", "share")
		}
		targetPath = filepath.Join(dataHome, "quickcast", "frecency.json")
	}

	s := &Store{
		filePath: targetPath,
		entries:  make(map[string]Entry),
	}

	if err := s.load(); err != nil && !os.IsNotExist(err) {
		// Log or return, but fallback to empty map
	}

	return s, nil
}

func (s *Store) load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.filePath)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, &s.entries)
}

// Save persists the frecency table to disk.
func (s *Store) Save() error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if err := os.MkdirAll(filepath.Dir(s.filePath), 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(s.entries, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.filePath, data, 0644)
}

// RecordAccess increments count and updates timestamp for an ID.
func (s *Store) RecordAccess(id string) {
	s.mu.Lock()
	entry := s.entries[id]
	entry.Count++
	entry.LastAccess = time.Now().Unix()
	s.entries[id] = entry
	s.mu.Unlock()

	_ = s.Save()
}

// FrecencyScore calculates score based on frequency and time decay.
// Similar to Mozilla / Raycast frecency algorithm:
// Recent within 4 hours: multiplier 100
// Within 24 hours: multiplier 70
// Within 3 days: multiplier 50
// Within 2 weeks: multiplier 30
// Older: multiplier 10
func (s *Store) FrecencyScore(id string) float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entry, ok := s.entries[id]
	if !ok || entry.Count == 0 {
		return 0
	}

	ageHours := float64(time.Now().Unix()-entry.LastAccess) / 3600.0

	var weight float64
	switch {
	case ageHours <= 4:
		weight = 100.0
	case ageHours <= 24:
		weight = 70.0
	case ageHours <= 72:
		weight = 50.0
	case ageHours <= 336: // 14 days
		weight = 30.0
	default:
		weight = 10.0
	}

	// Score = count * weight, log scaled to prevent runaway scores
	return math.Log1p(float64(entry.Count)) * weight
}
