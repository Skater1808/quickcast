#!/usr/bin/env bash

# @raycast.schemaVersion 1
# @raycast.title NixOS Rebuild Switch
# @raycast.mode compact
# @raycast.packageName NixOS
# @raycast.icon ❄️
# @raycast.description Rebuild NixOS system configuration live

if command -v notify-send >/dev/null; then
    notify-send "NixOS" "Starting system rebuild switch..." -i nixos
fi

if sudo nixos-rebuild switch; then
    if command -v notify-send >/dev/null; then
        notify-send "NixOS" "System rebuild switch succeeded! 🚀" -i nixos
    fi
else
    if command -v notify-send >/dev/null; then
        notify-send -u critical "NixOS" "Rebuild failed!" -i dialog-error
    fi
fi
