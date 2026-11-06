# ⚡ Quickcast

> **A Raycast replica built in Go for NixOS, Hyprland, and Quickshell.**

Quickcast recreates the beloved Raycast experience on Linux Wayland setups — combining the performance and safety of a Go daemon with the fluid GPU-accelerated layer-shell UI of Quickshell.

---

## 🎯 Architectural Analysis: Raycast vs. Quickcast

| Feature / Dimension | Raycast (macOS) | Quickcast (NixOS / Hyprland / Quickshell) |
| :--- | :--- | :--- |
| **Window Protocol** | Cocoa / AppKit floating overlay panel | Wayland `wlr-layer-shell-unstable-v1` via Quickshell |
| **Compositor FX** | macOS CoreAnimation backdrop blur | Hyprland `layerrule = blur, quickcast` + `ignorezero` |
| **Backend Core** | Native Swift / Rust core daemon | High-performance Go daemon (`quickcastd`) |
| **Frontend UI** | React / AppKit Node bridge | Hardware-accelerated declarative QML (QtQuick) |
| **IPC Mechanism** | Mach ports / XPC / UNIX sockets | High-throughput Unix Domain Socket with JSON streaming |
| **Search Engine** | Fuzzy search + Frecency scoring | Prefix, word-boundary, acronym fuzzy match + Frecency decay |
| **App Launcher** | macOS `.app` bundles & Spotlight index | Freedesktop `.desktop` files (NixOS profile, system, Flatpak) |
| **Window Switcher** | macOS WindowServer accessibility API | Hyprland IPC (`hyprctl -j clients` & `dispatch focuswindow`) |
| **Action Palette** | `Cmd+K` contextual action menu | `Ctrl+K` contextual action popup overlay |
| **Clipboard History** | In-memory + SQLite clipboard buffer | Wayland `wl-paste` watcher + searchable history |
| **Inline Calculator** | Natural math parser | Recursive-descent Pratt parser (`sqrt`, `sin`, `^`, constants) |
| **Script Commands** | `@raycast.*` metadata comment headers | Native parser for Raycast script metadata headers |
| **Packaging** | Shipped `.dmg` | Nix Flake, Home-Manager module, & Systemd user services |

---

## 🏗️ System Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                      Hyprland Compositor                    │
│   (Super+Space Hotkey, LayerShell Blur, Exclusive Focus)    │
└──────────────────────────────┬──────────────────────────────┘
                               │
               ┌───────────────┴───────────────┐
               ▼                               ▼
     ┌───────────────────┐           ┌───────────────────┐
     │  quickshell UI    │           │  quickcast (CLI)  │
     │   (shell.qml)     │           │ (toggle/run/ping) │
     └─────────┬─────────┘           └─────────┬─────────┘
               │                               │
               │    Unix Domain Socket (JSON)  │
               └───────────────┬───────────────┘
                               ▼
        ┌─────────────────────────────────────────────┐
        │          quickcastd (Go Core Daemon)        │
        │                                             │
        │  ┌───────────────────────────────────────┐  │
        │  │     Concurrent Search Engine          │  │
        │  └──────┬───────┬───────┬───────┬────────┘  │
        │         │       │       │       │           │
        │  ┌──────┴──┐ ┌──┴──┐ ┌──┴──┐ ┌──┴────────┐  │
        │  │  Apps   │ │Hypr │ │Math │ │Scripts/   │  │
        │  │(Desktop)│ │ctl  │ │Calc │ │Clipboard │  │
        │  └─────────┘ └─────┘ └─────┘ └───────────┘  │
        │                                             │
        │  ┌───────────────────────────────────────┐  │
        │  │      Frecency Scoring & History       │  │
        │  └───────────────────────────────────────┘  │
        └─────────────────────────────────────────────┘
```

---

## 🚀 Getting Started

### 1. NixOS / Flake Setup

Add Quickcast to your `flake.nix`:

```nix
{
  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
    quickcast.url = "github:emil/quickcast";
  };

  outputs = { self, nixpkgs, quickcast, ... }: {
    # In your Home-Manager configuration:
    homeManagerConfigurations."user" = home-manager.lib.homeManagerConfiguration {
      modules = [
        quickcast.homeManagerModules.default
        {
          programs.quickcast = {
            enable = true;
            hotkey = "$mainMod, SPACE"; # Super + Space
            autoStartHyprlandRules = true;
          };
        }
      ];
    };
  };
}
```

### 2. Manual Hyprland Configuration

If not using the Home-Manager module, add this to `~/.config/hypr/hyprland.conf`:

```ini
# Toggle Quickcast with Super + Space
bind = $mainMod, SPACE, exec, quickcast toggle

# Raycast-style Blur & Animations for Quickshell Layer
layerrule = blur, quickcast
layerrule = ignorezero, quickcast
layerrule = animation popin 80%, quickcast

# Autostart Daemon and UI
exec-once = quickcastd
exec-once = quickshell -p /path/to/quickcast/ui
```

### 3. Local Development & Testing

```bash
# Build both binaries
make build

# Run tests
make test

# Start the background daemon
make run-daemon

# Launch Quickshell interface
make run-ui
```

---

## ⌨️ Shortcuts & Controls

| Shortcut | Description |
| :--- | :--- |
| `Super + Space` | Toggle Quickcast launcher |
| `↑` / `↓` or `Ctrl+P` / `Ctrl+N` | Navigate search results |
| `↵` (Enter) | Execute primary action (Open app, Switch window, Copy result) |
| `Ctrl+K` | Open Contextual Action Palette (Raycast Action Menu) |
| `Ctrl+C` | Copy path or result to clipboard |
| `Esc` | Close Action Menu or dismiss Quickcast |

---

## 📜 Raycast Script Commands Compatibility

Quickcast natively runs any script with standard Raycast comment headers!
Drop scripts into `~/.config/quickcast/scripts/`:

```bash
#!/usr/bin/env bash

# @raycast.schemaVersion 1
# @raycast.title NixOS Rebuild Switch
# @raycast.mode compact
# @raycast.packageName NixOS
# @raycast.icon ❄️
# @raycast.description Rebuild NixOS live configuration

sudo nixos-rebuild switch && notify-send "NixOS" "Rebuild finished!"
```
Quickcast immediately indexes the script, displays it with its icon and package name, and runs it on selection.

---

## 🔌 Core Providers Included

1. **Applications**: Indexes `.desktop` files across `/run/current-system/sw`, `~/.nix-profile`, and Flatpaks.
2. **Hyprland Windows**: Real-time window switcher across workspaces via Hyprland IPC. Focus or kill windows.
3. **Inline Math / Calculator**: Instant evaluator supporting `+`, `-`, `*`, `/`, `^`, `%`, `sqrt()`, `sin()`, `abs()`, `pi`, `e`.
4. **Wayland Clipboard History**: Tracks copied text via `wl-paste` with fast search and paste back into active windows.
5. **System Controls**: Lock (`hyprlock`), Suspend, Reboot, Shutdown, and Hyprland Reload.
6. **Web Search Fallbacks**: Google, GitHub, and NixOS Packages (`search.nixos.org`).
