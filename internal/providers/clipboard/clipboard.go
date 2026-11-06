package clipboard

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/emil/quickcast/internal/search"
)

type ClipItem struct {
	ID        string
	Text      string
	Timestamp time.Time
}

type Provider struct {
	mu      sync.RWMutex
	history []ClipItem
	maxSize int
}

func New(maxSize int) *Provider {
	if maxSize <= 0 {
		maxSize = 100
	}
	p := &Provider{
		maxSize: maxSize,
		history: make([]ClipItem, 0, maxSize),
	}
	// Start background watcher if wl-paste is available
	go p.startWatcher()
	return p
}

func (p *Provider) Name() string {
	return "Clipboard History"
}

func (p *Provider) Refresh() error {
	return nil
}

func (p *Provider) Add(text string) {
	text = strings.TrimSpace(text)
	if text == "" {
		return
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	// Check if already top entry
	if len(p.history) > 0 && p.history[0].Text == text {
		return
	}

	// Filter out existing duplicates
	var filtered []ClipItem
	for _, it := range p.history {
		if it.Text != text {
			filtered = append(filtered, it)
		}
	}

	newItem := ClipItem{
		ID:        fmt.Sprintf("clip:%d", time.Now().UnixNano()),
		Text:      text,
		Timestamp: time.Now(),
	}

	p.history = append([]ClipItem{newItem}, filtered...)
	if len(p.history) > p.maxSize {
		p.history = p.history[:p.maxSize]
	}
}

func (p *Provider) startWatcher() {
	// Periodic check using wl-paste
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	var lastCopied string
	for range ticker.C {
		cmd := exec.Command("wl-paste", "--no-newline")
		var out bytes.Buffer
		cmd.Stdout = &out
		if err := cmd.Run(); err == nil {
			content := out.String()
			if content != "" && content != lastCopied {
				lastCopied = content
				p.Add(content)
			}
		}
	}
}

func (p *Provider) Search(query string) ([]search.Item, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var results []search.Item

	for _, item := range p.history {
		firstLine := item.Text
		if idx := strings.Index(firstLine, "\n"); idx != -1 {
			firstLine = firstLine[:idx]
		}
		if len(firstLine) > 60 {
			firstLine = firstLine[:60] + "..."
		}

		match := search.FuzzyMatch(query, item.Text)
		if query != "" && !match.Matched {
			continue
		}

		charCount := len([]rune(item.Text))
		timeAgo := formatTimeAgo(item.Timestamp)

		markdown := fmt.Sprintf("```\n%s\n```\n\n---\n- **Length:** %d chars\n- **Copied:** %s",
			item.Text, charCount, timeAgo)

		it := search.Item{
			ID:           item.ID,
			Title:        firstLine,
			Subtitle:     fmt.Sprintf("%d chars • %s", charCount, timeAgo),
			Category:     "Clipboard History",
			Icon:         "edit-paste",
			Accessories:  []string{"Clipboard"},
			Score:        match.Score,
			MatchIndices: match.Indices,
			Detail: &search.ItemDetail{
				Markdown: markdown,
				Metadata: map[string]string{
					"Length": fmt.Sprintf("%d chars", charCount),
					"Time":   timeAgo,
				},
			},
			Actions: []search.Action{
				{
					ID:       "copy",
					Title:    "Copy to Clipboard",
					Shortcut: "↵",
					Type:     "copy",
					Payload:  item.Text,
				},
				{
					ID:       "paste",
					Title:    "Paste to Active Window",
					Shortcut: "Ctrl+P",
					Type:     "paste",
					Payload:  item.Text,
				},
			},
		}

		results = append(results, it)
	}

	return results, nil
}

func formatTimeAgo(t time.Time) string {
	diff := time.Since(t)
	switch {
	case diff < time.Minute:
		return "just now"
	case diff < time.Hour:
		return fmt.Sprintf("%dm ago", int(diff.Minutes()))
	case diff < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(diff.Hours()))
	default:
		return fmt.Sprintf("%dd ago", int(diff.Hours()/24))
	}
}

func (p *Provider) Execute(action search.Action) error {
	switch action.Type {
	case "copy":
		cmd := exec.Command("wl-copy")
		cmd.Stdin = strings.NewReader(action.Payload)
		return cmd.Run()
	case "paste":
		// Copy first then simulate paste
		cmd := exec.Command("wl-copy")
		cmd.Stdin = strings.NewReader(action.Payload)
		if err := cmd.Run(); err != nil {
			return err
		}
		// Attempt ydotool or wtype for Wayland paste key event
		_ = exec.Command("wtype", "-M", "ctrl", "-k", "v", "-m", "ctrl").Run()
		return nil
	}
	return fmt.Errorf("unknown action: %s", action.Type)
}
