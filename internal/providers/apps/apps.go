package apps

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

type AppEntry struct {
	ID          string
	Name        string
	GenericName string
	Comment     string
	Exec        string
	Icon        string
	Path        string
	Terminal    bool
	Keywords    []string
	NoDisplay   bool
}

type Provider struct {
	mu      sync.RWMutex
	apps    []AppEntry
	appByID map[string]AppEntry
}

func New() *Provider {
	p := &Provider{
		appByID: make(map[string]AppEntry),
	}
	_ = p.Refresh()
	return p
}

func (p *Provider) Name() string {
	return "Applications"
}

// GetApplicationDirs returns all search directories for .desktop files,
// specially tailored for NixOS, Home-Manager, Flatpak, and Freedesktop standards.
func GetApplicationDirs() []string {
	var dirs []string
	home, _ := os.UserHomeDir()

	// 1. User local applications
	if home != "" {
		dirs = append(dirs,
			filepath.Join(home, ".local", "share", "applications"),
			filepath.Join(home, ".nix-profile", "share", "applications"),
			filepath.Join("/etc/profiles/per-user", os.Getenv("USER"), "share", "applications"),
		)
	}

	// 2. XDG_DATA_DIRS
	xdgDataDirs := os.Getenv("XDG_DATA_DIRS")
	if xdgDataDirs != "" {
		for _, dir := range strings.Split(xdgDataDirs, ":") {
			if dir != "" {
				dirs = append(dirs, filepath.Join(dir, "applications"))
			}
		}
	}

	// 3. NixOS system profile & standard Linux paths
	dirs = append(dirs,
		"/run/current-system/sw/share/applications",
		"/usr/local/share/applications",
		"/usr/share/applications",
		"/var/lib/flatpak/exports/share/applications",
	)

	// Deduplicate and filter non-existent dirs
	seen := make(map[string]bool)
	var validDirs []string
	for _, d := range dirs {
		cleaned := filepath.Clean(d)
		if !seen[cleaned] {
			seen[cleaned] = true
			if fi, err := os.Stat(cleaned); err == nil && fi.IsDir() {
				validDirs = append(validDirs, cleaned)
			}
		}
	}

	return validDirs
}

func (p *Provider) Refresh() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.apps = nil
	p.appByID = make(map[string]AppEntry)

	seenIDs := make(map[string]bool)
	dirs := GetApplicationDirs()

	for _, dir := range dirs {
		files, err := os.ReadDir(dir)
		if err != nil {
			continue
		}

		for _, file := range files {
			if file.IsDir() || !strings.HasSuffix(file.Name(), ".desktop") {
				continue
			}

			fullPath := filepath.Join(dir, file.Name())
			app, err := parseDesktopFile(fullPath)
			if err != nil || app.NoDisplay || app.Name == "" || app.Exec == "" {
				continue
			}

			// Desktop file ID convention
			id := "app:" + strings.TrimSuffix(file.Name(), ".desktop")
			if seenIDs[id] {
				continue
			}
			seenIDs[id] = true
			app.ID = id

			p.apps = append(p.apps, app)
			p.appByID[id] = app
		}
	}

	return nil
}

func parseDesktopFile(path string) (AppEntry, error) {
	file, err := os.Open(path)
	if err != nil {
		return AppEntry{}, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	inDesktopEntry := false
	var app AppEntry
	app.Path = path

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "[") {
			inDesktopEntry = (line == "[Desktop Entry]")
			continue
		}
		if !inDesktopEntry || strings.HasPrefix(line, "#") || !strings.Contains(line, "=") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])

		switch key {
		case "Name":
			if app.Name == "" {
				app.Name = val
			}
		case "GenericName":
			if app.GenericName == "" {
				app.GenericName = val
			}
		case "Comment":
			if app.Comment == "" {
				app.Comment = val
			}
		case "Exec":
			if app.Exec == "" {
				app.Exec = cleanExec(val)
			}
		case "Icon":
			if app.Icon == "" {
				app.Icon = val
			}
		case "Terminal":
			app.Terminal = (strings.ToLower(val) == "true")
		case "NoDisplay":
			app.NoDisplay = (strings.ToLower(val) == "true")
		case "Keywords":
			for _, kw := range strings.Split(val, ";") {
				kw = strings.TrimSpace(kw)
				if kw != "" {
					app.Keywords = append(app.Keywords, kw)
				}
			}
		}
	}

	return app, scanner.Err()
}

func cleanExec(execStr string) string {
	// Strip field codes (%f, %F, %u, %U, %d, %D, %n, %N, %i, %c, %k, %v, %m)
	fields := strings.Fields(execStr)
	var clean []string
	for _, f := range fields {
		if strings.HasPrefix(f, "%") {
			continue
		}
		clean = append(clean, f)
	}
	return strings.Join(clean, " ")
}

func (p *Provider) Search(query string) ([]search.Item, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var results []search.Item

	for _, app := range p.apps {
		// Fuzzy match name
		nameMatch := search.FuzzyMatch(query, app.Name)
		bestScore := nameMatch.Score
		var matchedIndices []int

		if nameMatch.Matched {
			matchedIndices = nameMatch.Indices
		}

		// Also check GenericName
		if app.GenericName != "" {
			gnMatch := search.FuzzyMatch(query, app.GenericName)
			if gnMatch.Matched && gnMatch.Score*0.8 > bestScore {
				bestScore = gnMatch.Score * 0.8
			}
		}

		// Also check Keywords
		for _, kw := range app.Keywords {
			kwMatch := search.FuzzyMatch(query, kw)
			if kwMatch.Matched && kwMatch.Score*0.7 > bestScore {
				bestScore = kwMatch.Score * 0.7
			}
		}

		if query != "" && !nameMatch.Matched && bestScore <= 0 {
			continue
		}

		sub := app.GenericName
		if sub == "" {
			sub = app.Comment
		}

		markdown := fmt.Sprintf("## %s\n\n%s\n\n---\n\n- **Command:** `%s`\n- **Path:** `%s`",
			app.Name, app.Comment, app.Exec, app.Path)

		metadata := map[string]string{
			"Name": app.Name,
			"Exec": app.Exec,
			"Path": app.Path,
		}
		if app.Terminal {
			metadata["Terminal"] = "Yes"
		}

		item := search.Item{
			ID:           app.ID,
			Title:        app.Name,
			Subtitle:     sub,
			Category:     "Applications",
			Icon:         app.Icon,
			Accessories:  []string{"Application"},
			Score:        bestScore,
			MatchIndices: matchedIndices,
			Detail: &search.ItemDetail{
				Markdown: markdown,
				Metadata: metadata,
			},
			Actions: []search.Action{
				{
					ID:       "open",
					Title:    "Open Application",
					Shortcut: "↵",
					Type:     "exec",
					Payload:  app.Exec,
				},
				{
					ID:       "copy_path",
					Title:    "Copy Desktop File Path",
					Shortcut: "Ctrl+C",
					Type:     "copy",
					Payload:  app.Path,
				},
				{
					ID:       "copy_cmd",
					Title:    "Copy Launch Command",
					Shortcut: "Ctrl+Y",
					Type:     "copy",
					Payload:  app.Exec,
				},
			},
		}

		results = append(results, item)
	}

	return results, nil
}

func (p *Provider) Execute(action search.Action) error {
	switch action.Type {
	case "exec":
		parts := strings.Fields(action.Payload)
		if len(parts) == 0 {
			return fmt.Errorf("empty command")
		}
		cmd := exec.Command(parts[0], parts[1:]...)
		cmd.Env = os.Environ()
		return cmd.Start()
	}
	return fmt.Errorf("unsupported action type: %s", action.Type)
}
