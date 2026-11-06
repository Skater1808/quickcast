import QtQuick
import QtQuick.Layouts

Rectangle {
    id: root

    property var itemData: null
    property bool isSelected: false
    property var theme: null

    signal clicked()
    signal doubleClicked()

    height: theme ? theme.itemHeight : 48
    radius: theme ? theme.radiusItem : 8
    color: isSelected ? (theme ? theme.bgSelected : "#363640") : "transparent"

    border.color: isSelected ? (theme ? theme.borderStrong : "#444450") : "transparent"
    border.width: 1

    Behavior on color {
        ColorAnimation { duration: 80 }
    }

    RowLayout {
        anchors.fill: parent
        anchors.leftMargin: 12
        anchors.rightMargin: 12
        spacing: 12

        // Left Icon
        Rectangle {
            id: iconContainer
            width: 32
            height: 32
            radius: 8
            color: root.isSelected ? (theme ? theme.bgElevated : "#222") : (theme ? theme.bgSurface : "#1a1a1a")

            Text {
                id: iconText
                anchors.centerIn: parent
                font.pixelSize: 18
                text: {
                    if (!itemData) return "⚡"
                    var cat = itemData.category || ""
                    if (cat === "Calculator") return "󰃬"
                    if (cat === "Hyprland Windows") return "󱂬"
                    if (cat === "Script Commands") return "󰆍"
                    if (cat === "Clipboard History") return "󰅌"
                    if (cat === "System") return "󰐥"
                    return "󰀻"
                }
                color: root.isSelected ? (theme ? theme.accent : "#ff6363") : (theme ? theme.textPrimary : "#fff")
            }
        }

        // Title and Subtitle
        ColumnLayout {
            Layout.fillWidth: true
            spacing: 2

            Text {
                Layout.fillWidth: true
                text: itemData ? itemData.title : ""
                font.family: theme ? theme.fontFamily : "sans-serif"
                font.pixelSize: 14
                font.weight: Font.Medium
                color: root.isSelected ? (theme ? theme.textPrimary : "#fff") : (theme ? theme.textPrimary : "#eee")
                elide: Text.ElideRight
            }

            Text {
                Layout.fillWidth: true
                visible: itemData && itemData.subtitle ? true : false
                text: itemData ? itemData.subtitle : ""
                font.family: theme ? theme.fontFamily : "sans-serif"
                font.pixelSize: 12
                color: theme ? theme.textSecondary : "#888"
                elide: Text.ElideRight
            }
        }

        // Accessories / Pills (Raycast badges)
        RowLayout {
            spacing: 6
            Repeater {
                model: (itemData && itemData.accessories) ? itemData.accessories : []
                Rectangle {
                    height: 20
                    radius: theme ? theme.radiusPill : 4
                    color: root.isSelected ? (theme ? theme.bgHover : "#282830") : (theme ? theme.bgSurface : "#1e1e24")
                    border.color: theme ? theme.borderSubtle : "#333"
                    border.width: 1

                    implicitWidth: accText.implicitWidth + 12

                    Text {
                        id: accText
                        anchors.centerIn: parent
                        text: modelData
                        font.family: theme ? theme.fontFamily : "sans-serif"
                        font.pixelSize: 10
                        font.weight: Font.Medium
                        color: theme ? theme.textMuted : "#aaa"
                    }
                }
            }
        }
    }

    MouseArea {
        anchors.fill: parent
        hoverEnabled: true
        onClicked: root.clicked()
        onDoubleClicked: root.doubleClicked()
    }
}
