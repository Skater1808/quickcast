package search

import (
	"fmt"
	"net/url"
	"os/exec"
	"sort"
	"strings"
	"sync"
)

type Engine struct {
	mu        sync.RWMutex
	providers []Provider
	frecency  FrecencyStore
}

type FrecencyStore interface {
	RecordAccess(id string)
	FrecencyScore(id string) float64
}

func NewEngine(frecency FrecencyStore) *Engine {
	return &Engine{
		frecency: frecency,
	}
}

func (e *Engine) Register(p Provider) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.providers = append(e.providers, p)
}

func (e *Engine) RefreshAll() {
	e.mu.RLock()
	defer e.mu.RUnlock()

	var wg sync.WaitGroup
	for _, p := range e.providers {
		wg.Add(1)
		go func(prov Provider) {
			defer wg.Done()
			_ = prov.Refresh()
		}(p)
	}
	wg.Wait()
}

func (e *Engine) Query(query string) []Item {
	e.mu.RLock()
	providers := make([]Provider, len(e.providers))
	copy(providers, e.providers)
	e.mu.RUnlock()

	var wg sync.WaitGroup
	var mu sync.Mutex
	var allItems []Item

	for _, p := range providers {
		wg.Add(1)
		go func(prov Provider) {
			defer wg.Done()
			items, err := prov.Search(query)
			if err != nil || len(items) == 0 {
				return
			}
			mu.Lock()
			allItems = append(allItems, items...)
			mu.Unlock()
		}(p)
	}

	wg.Wait()

	// Apply frecency scores
	for i := range allItems {
		if e.frecency != nil {
			fScore := e.frecency.FrecencyScore(allItems[i].ID)
			allItems[i].Score += fScore
		}
	}

	// Add Web Search fallbacks when query is non-empty
	if query != "" {
		encoded := url.QueryEscape(query)
		webFallbacks := []Item{
			{
				ID:          "fallback:google",
				Title:       fmt.Sprintf("Search Google for %q", query),
				Subtitle:    "Open search in default browser",
				Category:    "Web Search",
				Icon:        "web-browser",
				Accessories: []string{"Google"},
				Score:       1.0,
				Actions: []Action{
					{
						ID:       "open_google",
						Title:    "Search Google",
						Shortcut: "↵",
						Type:     "open_url",
						Payload:  "https://www.google.com/search?q=" + encoded,
					},
				},
			},
			{
				ID:          "fallback:nixos",
				Title:       fmt.Sprintf("Search NixOS Packages for %q", query),
				Subtitle:    "search.nixos.org",
				Category:    "Web Search",
				Icon:        "nix-snowflake",
				Accessories: []string{"NixOS"},
				Score:       0.9,
				Actions: []Action{
					{
						ID:       "open_nixos",
						Title:    "Search NixOS Packages",
						Shortcut: "↵",
						Type:     "open_url",
						Payload:  "https://search.nixos.org/packages?channel=unstable&query=" + encoded,
					},
				},
			},
			{
				ID:          "fallback:github",
				Title:       fmt.Sprintf("Search GitHub for %q", query),
				Subtitle:    "github.com/search",
				Category:    "Web Search",
				Icon:        "vcs-branch",
				Accessories: []string{"GitHub"},
				Score:       0.8,
				Actions: []Action{
					{
						ID:       "open_github",
						Title:    "Search GitHub",
						Shortcut: "↵",
						Type:     "open_url",
						Payload:  "https://github.com/search?q=" + encoded,
					},
				},
			},
		}
		allItems = append(allItems, webFallbacks...)
	}

	// Sort descending by score
	sort.SliceStable(allItems, func(i, j int) bool {
		return allItems[i].Score > allItems[j].Score
	})

	// Limit to top 50 results
	if len(allItems) > 50 {
		allItems = allItems[:50]
	}

	return allItems
}

func (e *Engine) Execute(action Action, itemID string) error {
	if e.frecency != nil && itemID != "" {
		e.frecency.RecordAccess(itemID)
	}

	switch action.Type {
	case "open_url":
		return exec.Command("xdg-open", action.Payload).Start()
	case "copy":
		cmd := exec.Command("wl-copy")
		cmd.Stdin = strings.NewReader(action.Payload)
		return cmd.Run()
	}

	// Find the matching provider
	e.mu.RLock()
	providers := make([]Provider, len(e.providers))
	copy(providers, e.providers)
	e.mu.RUnlock()

	for _, p := range providers {
		err := p.Execute(action)
		if err == nil {
			return nil
		}
	}

	return fmt.Errorf("no provider handled action: %s (%s)", action.ID, action.Type)
}
