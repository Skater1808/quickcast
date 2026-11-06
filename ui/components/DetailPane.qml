import QtQuick
import QtQuick.Layouts
import QtQuick.Controls

Rectangle {
    id: root

    property var itemData: null
    property var theme: null

    color: theme ? theme.bgSurface : "#18181c"
    radius: theme ? theme.radiusItem : 8
    border.color: theme ? theme.borderSubtle : "#2a2a32"
    border.width: 1

    ScrollView {
        anchors.fill: parent
        anchors.margins: 16
        clip: true

        ColumnLayout {
            width: parent.width
            spacing: 14

            // Header Banner
            RowLayout {
                Layout.fillWidth: true
                spacing: 12

                Rectangle {
                    width: 44
                    height: 44
                    radius: 10
                    color: theme ? theme.bgElevated : "#222"

                    Text {
                        anchors.centerIn: parent
                        font.pixelSize: 22
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
                        color: theme ? theme.accent : "#ff6363"
                    }
                }

                ColumnLayout {
                    Layout.fillWidth: true
                    spacing: 2

                    Text {
                        Layout.fillWidth: true
                        text: itemData ? itemData.title : ""
                        font.family: theme ? theme.fontFamily : "sans-serif"
                        font.pixelSize: 16
                        font.weight: Font.Bold
                        color: theme ? theme.textPrimary : "#fff"
                        elide: Text.ElideRight
                    }

                    Text {
                        Layout.fillWidth: true
                        text: itemData ? itemData.category : ""
                        font.family: theme ? theme.fontFamily : "sans-serif"
                        font.pixelSize: 12
                        color: theme ? theme.accent : "#ff6363"
                    }
                }
            }

            Rectangle {
                Layout.fillWidth: true
                height: 1
                color: theme ? theme.borderSubtle : "#333"
            }

            // Metadata key-values
            ColumnLayout {
                Layout.fillWidth: true
                spacing: 8
                visible: itemData && itemData.detail && itemData.detail.metadata

                Repeater {
                    model: {
                        if (!itemData || !itemData.detail || !itemData.detail.metadata) return []
                        var entries = []
                        for (var k in itemData.detail.metadata) {
                            entries.push({ key: k, value: itemData.detail.metadata[k] })
                        }
                        return entries
                    }

                    RowLayout {
                        Layout.fillWidth: true
                        spacing: 8

                        Text {
                            text: modelData.key
                            font.family: theme ? theme.fontFamily : "sans-serif"
                            font.pixelSize: 12
                            font.weight: Font.Medium
                            color: theme ? theme.textSecondary : "#888"
                            Layout.preferredWidth: 90
                        }

                        Text {
                            text: modelData.value
                            font.family: theme ? theme.fontMono : "monospace"
                            font.pixelSize: 11
                            color: theme ? theme.textPrimary : "#eee"
                            Layout.fillWidth: true
                            elide: Text.ElideMiddle
                        }
                    }
                }
            }

            // Markdown / Description Body
            Text {
                Layout.fillWidth: true
                visible: itemData && itemData.detail && itemData.detail.markdown ? true : false
                text: (itemData && itemData.detail && itemData.detail.markdown) ? itemData.detail.markdown : ""
                textFormat: Text.MarkdownText
                font.family: theme ? theme.fontFamily : "sans-serif"
                font.pixelSize: 13
                color: theme ? theme.textPrimary : "#ccc"
                wrapMode: Text.WordWrap
            }
        }
    }
}
