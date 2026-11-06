package scriptcommand

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/emil/quickcast/internal/search"
)

type ScriptCommand struct {
	Path        string
	Title       string
	PackageName string
	Icon        string
	Mode        string // "compact", "fullOutput", "silent"
	Description string
	Author      string
}

type Provider struct {
	mu      sync.RWMutex
	scripts []ScriptCommand
	dirs    []string
}

func New(customDirs ...string) *Provider {
	var dirs []string
	if len(customDirs) > 0 {
		dirs = customDirs
	} else {
		home, _ := os.UserHomeDir()
		if home != "" {
			dirs = append(dirs, filepath.Join(home, ".config", "quickcast", "scripts"))
		}
	}

	p := &Provider{dirs: dirs}
	_ = p.Refresh()
	return p
}

func (p *Provider) Name() string {
	return "Script Commands"
}

func (p *Provider) Refresh() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.scripts = nil

	for _, dir := range p.dirs {
		files, err := os.ReadDir(dir)
		if err != nil {
			continue
		}

		for _, f := range files {
			if f.IsDir() {
				continue
			}

			fullPath := filepath.Join(dir, f.Name())
			fi, err := f.Info()
			if err != nil || (fi.Mode()&0111) == 0 {
				// Must be executable
				continue
			}

			cmd, err := parseScript(fullPath)
			if err != nil || cmd.Title == "" {
				continue
			}

			p.scripts = append(p.scripts, cmd)
		}
	}

	return nil
}

func parseScript(path string) (ScriptCommand, error) {
	file, err := os.Open(path)
	if err != nil {
		return ScriptCommand{}, err
	}
	defer file.Close()

	cmd := ScriptCommand{
		Path:        path,
		Title:       filepath.Base(path),
		PackageName: "Scripts",
		Mode:        "compact",
		Icon:        "utilities-terminal",
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "#") {
			// Stop scanning if past header comments
			if len(line) > 0 && !strings.HasPrefix(line, "#!") {
				break
			}
			continue
		}

		line = strings.TrimPrefix(line, "#")
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "@raycast.") || strings.HasPrefix(line, "@quickcast.") {
			parts := strings.SplitN(line, " ", 2)
			tag := strings.ToLower(parts[0])
			val := ""
			if len(parts) > 1 {
				val = strings.TrimSpace(parts[1])
			}

			switch tag {
			case "@raycast.title", "@quickcast.title":
				cmd.Title = val
			case "@raycast.packagename", "@quickcast.packagename":
				cmd.PackageName = val
			case "@raycast.icon", "@quickcast.icon":
				cmd.Icon = val
			case "@raycast.mode", "@quickcast.mode":
				cmd.Mode = val
			case "@raycast.description", "@quickcast.description":
				cmd.Description = val
			case "@raycast.author", "@quickcast.author":
				cmd.Author = val
			}
		}
	}

	return cmd, scanner.Err()
}

func (p *Provider) Search(query string) ([]search.Item, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var results []search.Item

	for _, sc := range p.scripts {
		titleMatch := search.FuzzyMatch(query, sc.Title)
		pkgMatch := search.FuzzyMatch(query, sc.PackageName)

		bestScore := titleMatch.Score
		indices := titleMatch.Indices

		if pkgMatch.Score*0.7 > bestScore {
			bestScore = pkgMatch.Score * 0.7
		}

		if query != "" && !titleMatch.Matched && bestScore <= 0 {
			continue
		}

		markdown := fmt.Sprintf("## %s\n\n%s\n\n---\n- **Package:** `%s`\n- **Mode:** `%s`\n- **Path:** `%s`",
			sc.Title, sc.Description, sc.PackageName, sc.Mode, sc.Path)

		metadata := map[string]string{
			"Package": sc.PackageName,
			"Mode":    sc.Mode,
			"Path":    sc.Path,
		}
		if sc.Author != "" {
			metadata["Author"] = sc.Author
		}

		item := search.Item{
			ID:           "script:" + sc.Path,
			Title:        sc.Title,
			Subtitle:     fmt.Sprintf("%s • %s", sc.PackageName, sc.Description),
			Category:     "Script Commands",
			Icon:         sc.Icon,
			Accessories:  []string{sc.PackageName, "Script"},
			Score:        bestScore,
			MatchIndices: indices,
			Detail: &search.ItemDetail{
				Markdown: markdown,
				Metadata: metadata,
			},
			Actions: []search.Action{
				{
					ID:       "run",
					Title:    "Run Script",
					Shortcut: "↵",
					Type:     "script",
					Payload:  sc.Path,
				},
				{
					ID:       "edit",
					Title:    "Edit Script",
					Shortcut: "Ctrl+E",
					Type:     "exec",
					Payload:  fmt.Sprintf("$EDITOR %s", sc.Path),
				},
				{
					ID:       "copy_path",
					Title:    "Copy Script Path",
					Shortcut: "Ctrl+C",
					Type:     "copy",
					Payload:  sc.Path,
				},
			},
		}

		results = append(results, item)
	}

	return results, nil
}

func (p *Provider) Execute(action search.Action) error {
	switch action.Type {
	case "script":
		cmd := exec.Command(action.Payload)
		cmd.Env = os.Environ()
		return cmd.Start()
	case "exec":
		parts := strings.Fields(action.Payload)
		if len(parts) == 0 {
			return nil
		}
		cmd := exec.Command(parts[0], parts[1:]...)
		return cmd.Start()
	}
	return fmt.Errorf("unsupported action: %s", action.Type)
}
