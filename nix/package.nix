{ lib
, buildGoModule
, makeWrapper
, wl-clipboard
, wtype
, quickshell
, hyprland
}:

buildGoModule {
  pname = "quickcast";
  version = "0.1.0";

  src = ../.;

  vendorHash = null; # Pure Go standard library dependencies

  nativeBuildInputs = [ makeWrapper ];

  subPackages = [ "cmd/quickcast" "cmd/quickcastd" ];

  postInstall = ''
    wrapProgram $out/bin/quickcast \
      --prefix PATH : ${lib.makeBinPath [ quickshell hyprland wl-clipboard wtype ]}

    wrapProgram $out/bin/quickcastd \
      --prefix PATH : ${lib.makeBinPath [ quickshell hyprland wl-clipboard wtype ]}

    # Install quickshell UI assets
    mkdir -p $out/share/quickcast/ui
    cp -r $src/ui/* $out/share/quickcast/ui/
    cp -r $src/scripts $out/share/quickcast/
  '';

  meta = with lib; {
    description = "Raycast-inspired modern extensible launcher for NixOS, Hyprland, and Quickshell";
    homepage = "https://github.com/emil/quickcast";
    license = licenses.mit;
    platforms = platforms.linux;
    maintainers = [ ];
  };
}
