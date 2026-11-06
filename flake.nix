{
  description = "Quickcast: Raycast reconstructed in Go for NixOS, Hyprland & Quickshell";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs { inherit system; };
      in
      {
        packages.default = pkgs.callPackage ./nix/package.nix { };

        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            go
            gopls
            wl-clipboard
            wtype
            hyprland
          ];
        };
      }) // {
        homeManagerModules.default = import ./nix/module.nix;
        nixosModules.default = import ./nix/module.nix;
      };
}
