package system

import (
	"fmt"
	"os/exec"

	"github.com/emil/quickcast/internal/search"
)

type SystemCommand struct {
	ID          string
	Title       string
	Subtitle    string
	Icon        string
	Command     string
	Keywords    []string
	Description string
}

type Provider struct {
	commands []SystemCommand
}

func New() *Provider {
	return &Provider{
		commands: []SystemCommand{
			{
				ID:          "sys:lock",
				Title:       "Lock Screen",
				Subtitle:    "Lock current Wayland session with hyprlock",
				Icon:        "system-lock-screen",
				Command:     "hyprlock || loginctl lock-session",
				Keywords:    []string{"lock", "screen", "hyprlock"},
				Description: "Locks the display using hyprlock or loginctl.",
			},
			{
				ID:          "sys:suspend",
				Title:       "Sleep / Suspend",
				Subtitle:    "Suspend system to RAM",
				Icon:        "system-suspend",
				Command:     "systemctl suspend",
				Keywords:    []string{"sleep", "suspend", "standby"},
				Description: "Suspends the machine via systemd.",
			},
			{
				ID:          "sys:reboot",
				Title:       "Restart System",
				Subtitle:    "Reboot the computer",
				Icon:        "system-reboot",
				Command:     "systemctl reboot",
				Keywords:    []string{"reboot", "restart"},
				Description: "Restarts the system via systemd.",
			},
			{
				ID:          "sys:poweroff",
				Title:       "Shut Down",
				Subtitle:    "Power off the computer",
				Icon:        "system-shutdown",
				Command:     "systemctl poweroff",
				Keywords:    []string{"shutdown", "poweroff", "turn off"},
				Description: "Powers down the computer.",
			},
			{
				ID:          "sys:hypr_reload",
				Title:       "Reload Hyprland Config",
				Subtitle:    "hyprctl reload",
				Icon:        "view-refresh",
				Command:     "hyprctl reload",
				Keywords:    []string{"hyprland", "reload", "config"},
				Description: "Reloads Hyprland configuration live.",
			},
		},
	}
}

func (p *Provider) Name() string {
	return "System Commands"
}

func (p *Provider) Refresh() error {
	return nil
}

func (p *Provider) Search(query string) ([]search.Item, error) {
	var results []search.Item

	for _, cmd := range p.commands {
		match := search.FuzzyMatch(query, cmd.Title)
		bestScore := match.Score
		indices := match.Indices

		for _, kw := range cmd.Keywords {
			kwMatch := search.FuzzyMatch(query, kw)
			if kwMatch.Score*0.8 > bestScore {
				bestScore = kwMatch.Score * 0.8
			}
		}

		if query != "" && !match.Matched && bestScore <= 0 {
			continue
		}

		markdown := fmt.Sprintf("## %s\n\n%s\n\n---\n- **Command:** `%s`", cmd.Title, cmd.Description, cmd.Command)

		it := search.Item{
			ID:           cmd.ID,
			Title:        cmd.Title,
			Subtitle:     cmd.Subtitle,
			Category:     "System",
			Icon:         cmd.Icon,
			Accessories:  []string{"System Command"},
			Score:        bestScore,
			MatchIndices: indices,
			Detail: &search.ItemDetail{
				Markdown: markdown,
				Metadata: map[string]string{
					"Command": cmd.Command,
				},
			},
			Actions: []search.Action{
				{
					ID:       "run",
					Title:    cmd.Title,
					Shortcut: "↵",
					Type:     "shell",
					Payload:  cmd.Command,
				},
			},
		}

		results = append(results, it)
	}

	return results, nil
}

func (p *Provider) Execute(action search.Action) error {
	if action.Type == "shell" {
		cmd := exec.Command("sh", "-c", action.Payload)
		return cmd.Start()
	}
	return fmt.Errorf("unknown action: %s", action.Type)
}
