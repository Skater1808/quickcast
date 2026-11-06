package search

// Action represents an executable operation associated with an item.
type Action struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Shortcut string `json:"shortcut,omitempty"` // e.g. "Enter", "Ctrl+Enter", "Ctrl+Y"
	Type     string `json:"type"`               // "exec", "copy", "open_url", "hyprctl", "script"
	Payload  string `json:"payload"`
}

// ItemDetail provides markdown preview or structured metadata for the Raycast detail pane.
type ItemDetail struct {
	Markdown string            `json:"markdown,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
	Preview  string            `json:"preview,omitempty"`
}

// Item is a search result item shown in the launcher.
type Item struct {
	ID          string            `json:"id"`
	Title       string            `json:"title"`
	Subtitle    string            `json:"subtitle,omitempty"`
	Category    string            `json:"category"` // "Applications", "Hyprland Windows", "Calculator", "Script Commands", "Clipboard", "System"
	Icon        string            `json:"icon"`     // Freedesktop icon name, file path, nerd font or emoji
	Accessories []string          `json:"accessories,omitempty"`
	Keywords    []string          `json:"keywords,omitempty"`
	Score       float64           `json:"score,omitempty"`
	MatchIndices []int            `json:"match_indices,omitempty"`
	Detail      *ItemDetail       `json:"detail,omitempty"`
	Actions     []Action          `json:"actions"`
}

// Provider defines the interface for sources of items and search results.
type Provider interface {
	Name() string
	// Search queries the provider with the user string.
	Search(query string) ([]Item, error)
	// Refresh reloads or re-indexes cached resources if applicable.
	Refresh() error
	// Execute performs an action for this provider's item.
	Execute(action Action) error
}
