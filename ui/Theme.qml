import QtQuick

QtObject {
    id: root

    // Palette inspired by Raycast Dark
    property color bgBase: "#141417"
    property color bgSurface: "#1c1c21"
    property color bgElevated: "#232329"
    property color bgHover: "#2c2c34"
    property color bgSelected: "#363640"
    
    // Raycast signature red accent
    property color accent: "#ff6363"
    property color accentHover: "#ff7a7a"
    property color accentMuted: "#3b1e22"

    // Text colors
    property color textPrimary: "#f5f5f7"
    property color textSecondary: "#8e8e93"
    property color textMuted: "#5e5e66"
    property color textHighlight: "#ff6363"

    // Borders & Dividers
    property color borderSubtle: "#2d2d35"
    property color borderStrong: "#3d3d47"

    // Radii
    property real radiusWindow: 14
    property real radiusItem: 8
    property real radiusPill: 6

    // Dimensions
    property real windowWidth: 840
    property real windowHeight: 520
    property real searchBarHeight: 64
    property real itemHeight: 48
    property real statusBarHeight: 38

    // Fonts
    property string fontFamily: "Inter, Cantarell, Ubuntu, sans-serif"
    property string fontMono: "JetBrains Mono, Fira Code, monospace"
}
