import QtQuick
import QtQuick.Layouts

Rectangle {
    id: root

    property var actions: []
    property int selectedIndex: 0
    property var theme: null

    signal actionTriggered(var action)
    signal closeRequested()

    color: theme ? theme.bgBase : "#141417"
    border.color: theme ? theme.borderStrong : "#3d3d47"
    border.width: 1
    radius: theme ? theme.radiusWindow : 12

    ColumnLayout {
        anchors.fill: parent
        anchors.margins: 12
        spacing: 8

        // Header
        RowLayout {
            Layout.fillWidth: true
            spacing: 8

            Text {
                text: "Actions"
                font.family: theme ? theme.fontFamily : "sans-serif"
                font.pixelSize: 13
                font.weight: Font.DemiBold
                color: theme ? theme.textSecondary : "#888"
                Layout.fillWidth: true
            }

            Rectangle {
                height: 18
                radius: 4
                color: theme ? theme.bgElevated : "#222"
                implicitWidth: escText.implicitWidth + 8

                Text {
                    id: escText
                    anchors.centerIn: parent
                    text: "Esc"
                    font.pixelSize: 10
                    color: theme ? theme.textMuted : "#666"
                }
            }
        }

        Rectangle {
            Layout.fillWidth: true
            height: 1
            color: theme ? theme.borderSubtle : "#333"
        }

        // Action List
        ListView {
            id: actionList
            Layout.fillWidth: true
            Layout.fillHeight: true
            clip: true
            model: root.actions
            currentIndex: root.selectedIndex

            delegate: Rectangle {
                width: actionList.width
                height: 38
                radius: theme ? theme.radiusItem : 6
                color: index === actionList.currentIndex ? (theme ? theme.bgSelected : "#363640") : "transparent"

                RowLayout {
                    anchors.fill: parent
                    anchors.leftMargin: 10
                    anchors.rightMargin: 10
                    spacing: 8

                    Text {
                        text: modelData.title
                        font.family: theme ? theme.fontFamily : "sans-serif"
                        font.pixelSize: 13
                        color: index === actionList.currentIndex ? (theme ? theme.textPrimary : "#fff") : (theme ? theme.textPrimary : "#ddd")
                        Layout.fillWidth: true
                    }

                    Rectangle {
                        visible: modelData.shortcut ? true : false
                        height: 20
                        radius: 4
                        color: theme ? theme.bgElevated : "#2a2a30"
                        implicitWidth: shortcutText.implicitWidth + 8

                        Text {
                            id: shortcutText
                            anchors.centerIn: parent
                            text: modelData.shortcut ? modelData.shortcut : ""
                            font.family: theme ? theme.fontMono : "monospace"
                            font.pixelSize: 10
                            font.weight: Font.Medium
                            color: theme ? theme.textSecondary : "#aaa"
                        }
                    }
                }

                MouseArea {
                    anchors.fill: parent
                    onClicked: {
                        actionList.currentIndex = index
                        root.actionTriggered(modelData)
                    }
                }
            }
        }
    }

    function selectNext() {
        if (actionList.count > 0) {
            actionList.currentIndex = (actionList.currentIndex + 1) % actionList.count
        }
    }

    function selectPrev() {
        if (actionList.count > 0) {
            actionList.currentIndex = (actionList.currentIndex - 1 + actionList.count) % actionList.count
        }
    }

    function triggerCurrent() {
        if (actionList.currentIndex >= 0 && actionList.currentIndex < root.actions.length) {
            root.actionTriggered(root.actions[actionList.currentIndex])
        }
    }
}
