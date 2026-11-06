package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/emil/quickcast/internal/config"
	"github.com/emil/quickcast/internal/ipc"
	"github.com/emil/quickcast/internal/providers/apps"
	"github.com/emil/quickcast/internal/providers/calculator"
	"github.com/emil/quickcast/internal/providers/clipboard"
	"github.com/emil/quickcast/internal/providers/hyprland"
	"github.com/emil/quickcast/internal/providers/scriptcommand"
	"github.com/emil/quickcast/internal/providers/system"
	"github.com/emil/quickcast/internal/search"
	"github.com/emil/quickcast/internal/state"
)

type appHandler struct {
	engine *search.Engine
}

func (h *appHandler) OnSearch(query string) []search.Item {
	return h.engine.Query(query)
}

func (h *appHandler) OnExecute(itemID string, action search.Action) error {
	return h.engine.Execute(action, itemID)
}

func main() {
	socketFlag := flag.String("socket", "", "Custom unix socket path")
	flag.Parse()

	log.Println("🚀 Starting Quickcast Daemon (Raycast for NixOS / Hyprland)...")

	cfg := config.LoadConfig()

	// 1. Initialize Frecency State Store
	frecencyStore, err := state.NewStore()
	if err != nil {
		log.Printf("Warning: failed to load frecency store: %v", err)
	}

	// 2. Initialize Search Engine
	engine := search.NewEngine(frecencyStore)

	// 3. Register Providers
	engine.Register(apps.New())
	engine.Register(calculator.New())
	engine.Register(system.New())

	if cfg.EnableHyprland {
		engine.Register(hyprland.New())
	}

	if cfg.EnableClipboard {
		engine.Register(clipboard.New(100))
	}

	engine.Register(scriptcommand.New(cfg.ScriptDirs...))

	// Initial scan
	engine.RefreshAll()

	// 4. Start IPC Server
	handler := &appHandler{engine: engine}
	server := ipc.NewServer(handler, *socketFlag)
	if err := server.Start(); err != nil {
		log.Fatalf("Fatal: failed to start IPC server: %v", err)
	}
	defer server.Stop()

	log.Printf("⚡ Quickcast Daemon listening on %s", ipc.GetSocketPath())

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down Quickcast Daemon...")
}
