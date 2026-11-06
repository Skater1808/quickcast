package hyprland

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/emil/quickcast/internal/search"
)

type Client struct {
	Address        string `json:"address"`
	At             [2]int `json:"at"`
	Size           [2]int `json:"size"`
	Workspace      Workspace `json:"workspace"`
	Floating       bool   `json:"floating"`
	Monitor        int    `json:"monitor"`
	Class          string `json:"class"`
	Title          string `json:"title"`
	InitialClass   string `json:"initialClass"`
	InitialTitle   string `json:"initialTitle"`
	Pid            int    `json:"pid"`
	Xwayland       bool   `json:"xwayland"`
	Pinned         bool   `json:"pinned"`
	Fullscreen     bool   `json:"fullscreen"`
	FullscreenMode int    `json:"fullscreenMode"`
}

type Workspace struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Provider struct{}

func New() *Provider {
	return &Provider{}
}

func (p *Provider) Name() string {
	return "Hyprland Windows"
}

func (p *Provider) IsAvailable() bool {
	return os.Getenv("HYPRLAND_INSTANCE_SIGNATURE") != ""
}

func (p *Provider) Refresh() error {
	return nil
}

func (p *Provider) getClients() ([]Client, error) {
	if !p.IsAvailable() {
		return nil, nil
	}

	cmd := exec.Command("hyprctl", "-j", "clients")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var clients []Client
	if err := json.Unmarshal(out, &clients); err != nil {
		return nil, err
	}

	return clients, nil
}

func (p *Provider) Search(query string) ([]search.Item, error) {
	clients, err := p.getClients()
	if err != nil || len(clients) == 0 {
		return nil, nil
	}

	var results []search.Item

	for _, c := range clients {
		// Skip empty titles or system windows
		if c.Title == "" && c.Class == "" {
			continue
		}

		titleMatch := search.FuzzyMatch(query, c.Title)
		classMatch := search.FuzzyMatch(query, c.Class)

		bestScore := titleMatch.Score
		indices := titleMatch.Indices

		if classMatch.Score > bestScore {
			bestScore = classMatch.Score
			indices = classMatch.Indices
		}

		if query != "" && !titleMatch.Matched && !classMatch.Matched {
			continue
		}

		// Bonus score for active Hyprland windows
		bestScore += 20.0

		wsName := c.Workspace.Name
		if wsName == "" {
			wsName = fmt.Sprintf("Workspace %d", c.Workspace.ID)
		}

		markdown := fmt.Sprintf("## %s\n\n- **Class:** `%s`\n- **Workspace:** `%s`\n- **Address:** `%s`\n- **PID:** `%d`\n- **Floating:** `%v`",
			c.Title, c.Class, wsName, c.Address, c.Pid, c.Floating)

		metadata := map[string]string{
			"Class":     c.Class,
			"Workspace": wsName,
			"PID":       fmt.Sprintf("%d", c.Pid),
			"Address":   c.Address,
		}

		item := search.Item{
			ID:           "hypr:" + c.Address,
			Title:        c.Title,
			Subtitle:     fmt.Sprintf("%s • %s", c.Class, wsName),
			Category:     "Hyprland Windows",
			Icon:         strings.ToLower(c.Class),
			Accessories:  []string{wsName, "Window"},
			Score:        bestScore,
			MatchIndices: indices,
			Detail: &search.ItemDetail{
				Markdown: markdown,
				Metadata: metadata,
			},
			Actions: []search.Action{
				{
					ID:       "focus",
					Title:    "Focus Window",
					Shortcut: "↵",
					Type:     "hyprctl",
					Payload:  fmt.Sprintf("dispatch focuswindow address:%s", c.Address),
				},
				{
					ID:       "close",
					Title:    "Close Window",
					Shortcut: "Ctrl+W",
					Type:     "hyprctl",
					Payload:  fmt.Sprintf("dispatch closewindow address:%s", c.Address),
				},
				{
					ID:       "toggle_float",
					Title:    "Toggle Floating",
					Shortcut: "Ctrl+F",
					Type:     "hyprctl",
					Payload:  fmt.Sprintf("dispatch togglefloating address:%s", c.Address),
				},
				{
					ID:       "kill",
					Title:    "Force Kill Window",
					Shortcut: "Ctrl+K",
					Type:     "hyprctl",
					Payload:  fmt.Sprintf("dispatch killactive"),
				},
			},
		}

		results = append(results, item)
	}

	return results, nil
}

func (p *Provider) Execute(action search.Action) error {
	if action.Type == "hyprctl" {
		args := strings.Fields(action.Payload)
		cmd := exec.Command("hyprctl", args...)
		return cmd.Run()
	}
	return fmt.Errorf("unknown action type: %s", action.Type)
}
