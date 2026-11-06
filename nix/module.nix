{ config, lib, pkgs, ... }:

with lib;

let
  cfg = config.programs.quickcast;
  defaultPackage = pkgs.callPackage ./package.nix { };
in {
  options.programs.quickcast = {
    enable = mkEnableOption "Quickcast - Raycast-style launcher for Hyprland and NixOS";

    package = mkOption {
      type = types.package;
      default = defaultPackage;
      description = "The Quickcast package to use.";
    };

    hotkey = mkOption {
      type = types.str;
      default = "$mainMod, SPACE";
      description = "Hyprland key combination to trigger Quickcast toggle.";
    };

    autoStartHyprlandRules = mkOption {
      type = types.bool;
      default = true;
      description = "Whether to inject Hyprland layer rules for Quickcast (blur, animation).";
    };
  };

  config = mkIf cfg.enable {
    home.packages = [ cfg.package ];

    # Systemd User Unit for Quickcast Daemon
    systemd.user.services.quickcastd = {
      Unit = {
        Description = "Quickcast Backend Daemon";
        PartOf = [ "graphical-session.target" ];
        After = [ "graphical-session.target" ];
      };
      Service = {
        ExecStart = "${cfg.package}/bin/quickcastd";
        Restart = "on-failure";
        RestartSec = 3;
      };
      Install = {
        WantedBy = [ "graphical-session.target" ];
      };
    };

    # Systemd User Unit for Quickshell UI
    systemd.user.services.quickcast-ui = {
      Unit = {
        Description = "Quickcast Quickshell UI";
        Requires = [ "quickcastd.service" ];
        After = [ "quickcastd.service" "graphical-session.target" ];
      };
      Service = {
        ExecStart = "${pkgs.quickshell}/bin/quickshell -p ${cfg.package}/share/quickcast/ui";
        Restart = "on-failure";
        RestartSec = 3;
      };
      Install = {
        WantedBy = [ "graphical-session.target" ];
      };
    };

    # Inject Hyprland Configuration if wayland.windowManager.hyprland is enabled
    wayland.windowManager.hyprland.settings = mkIf cfg.autoStartHyprlandRules {
      bind = [
        "${cfg.hotkey}, exec, ${cfg.package}/bin/quickcast toggle"
      ];
      layerrule = [
        "blur, quickcast"
        "ignorezero, quickcast"
        "animation popin 80%, quickcast"
      ];
    };
  };
}
