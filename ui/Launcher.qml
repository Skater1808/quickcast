import QtQuick
import QtQuick.Layouts
import QtQuick.Controls
import "components"

Rectangle {
    id: root

    property var theme: Theme {}
    property var items: []
    property int selectedIndex: 0
    property bool isActionMenuOpen: false

    signal queryChanged(string text)
    signal executeRequested(string itemId, var action)
    signal closeRequested()

    width: theme.windowWidth
    height: theme.windowHeight
    radius: theme.radiusWindow
    color: theme.bgBase
    border.color: theme.borderStrong
    border.width: 1
    clip: true

    // Top Search Bar
    Rectangle {
        id: searchBar
        anchors.top: parent.top
        anchors.left: parent.left
        anchors.right: parent.right
        height: theme.searchBarHeight
        color: "transparent"

        RowLayout {
            anchors.fill: parent
            anchors.leftMargin: 20
            anchors.rightMargin: 20
            spacing: 14

            Text {
                text: ""
                font.pixelSize: 18
                color: searchInput.text.length > 0 ? theme.accent : theme.textMuted
            }

            TextInput {
                id: searchInput
                Layout.fillWidth: true
                font.family: theme.fontFamily
                font.pixelSize: 18
                color: theme.textPrimary
                selectionColor: theme.accent
                selectedTextColor: "#ffffff"
                focus: true

                Text {
                    anchors.fill: parent
                    text: "Search for applications, commands, or math..."
                    font.family: searchInput.font.family
                    font.pixelSize: searchInput.font.pixelSize
                    color: theme.textMuted
                    visible: !searchInput.text && !searchInput.inputMethodComposing
                }

                onTextChanged: {
                    debounceTimer.restart()
                }

                Keys.onPressed: function(event) {
                    if (event.key === Qt.Key_Escape) {
                        if (root.isActionMenuOpen) {
                            root.isActionMenuOpen = false
                            event.accepted = true
                        } else {
                            root.closeRequested()
                            event.accepted = true
                        }
                    } else if (event.key === Qt.Key_Down || (event.modifiers & Qt.ControlModifier && event.key === Qt.Key_N)) {
                        if (root.isActionMenuOpen) {
                            actionMenu.selectNext()
                        } else {
                            selectNextItem()
                        }
                        event.accepted = true
                    } else if (event.key === Qt.Key_Up || (event.modifiers & Qt.ControlModifier && event.key === Qt.Key_P)) {
                        if (root.isActionMenuOpen) {
                            actionMenu.selectPrev()
                        } else {
                            selectPrevItem()
                        }
                        event.accepted = true
                    } else if (event.key === Qt.Key_Return || event.key === Qt.Key_Enter) {
                        if (root.isActionMenuOpen) {
                            actionMenu.triggerCurrent()
                        } else {
                            triggerSelectedItem()
                        }
                        event.accepted = true
                    } else if (event.modifiers & Qt.ControlModifier && event.key === Qt.Key_K) {
                        toggleActionMenu()
                        event.accepted = true
                    }
                }
            }

            // Clear search button
            Rectangle {
                width: 22
                height: 22
                radius: 11
                color: theme.bgElevated
                visible: searchInput.text.length > 0

                Text {
                    anchors.centerIn: parent
                    text: "✕"
                    font.pixelSize: 11
                    color: theme.textSecondary
                }

                MouseArea {
                    anchors.fill: parent
                    onClicked: {
                        searchInput.text = ""
                        searchInput.forceActiveFocus()
                    }
                }
            }
        }

        Rectangle {
            anchors.bottom: parent.bottom
            anchors.left: parent.left
            anchors.right: parent.right
            height: 1
            color: theme.borderSubtle
        }
    }

    // Debounce timer for smooth typing without lag
    Timer {
        id: debounceTimer
        interval: 80
        repeat: false
        onTriggered: {
            root.queryChanged(searchInput.text)
        }
    }

    // Split Content Body: Results List on Left, Detail on Right
    RowLayout {
        anchors.top: searchBar.bottom
        anchors.bottom: statusBar.top
        anchors.left: parent.left
        anchors.right: parent.right
        anchors.margins: 12
        spacing: 12

        // Left Results List
        ListView {
            id: resultsList
            Layout.fillWidth: true
            Layout.fillHeight: true
            clip: true
            model: root.items
            currentIndex: root.selectedIndex

            delegate: ItemRow {
                width: resultsList.width
                itemData: modelData
                theme: root.theme
                isSelected: index === resultsList.currentIndex

                onClicked: {
                    resultsList.currentIndex = index
                    root.selectedIndex = index
                }

                onDoubleClicked: {
                    root.triggerSelectedItem()
                }
            }

            Text {
                anchors.centerIn: parent
                visible: resultsList.count === 0 && searchInput.text.length > 0
                text: "No results found"
                font.family: theme.fontFamily
                font.pixelSize: 14
                color: theme.textMuted
            }
        }

        // Right Detail Pane (visible when active item has detail)
        DetailPane {
            id: detailPane
            Layout.preferredWidth: parent.width * 0.44
            Layout.fillHeight: true
            theme: root.theme
            visible: currentItem() && currentItem().detail ? true : false
            itemData: currentItem()
        }
    }

    // Bottom Status Bar
    Rectangle {
        id: statusBar
        anchors.bottom: parent.bottom
        anchors.left: parent.left
        anchors.right: parent.right
        height: theme.statusBarHeight
        color: theme.bgSurface

        Rectangle {
            anchors.top: parent.top
            anchors.left: parent.left
            anchors.right: parent.right
            height: 1
            color: theme.borderSubtle
        }

        RowLayout {
            anchors.fill: parent
            anchors.leftMargin: 16
            anchors.rightMargin: 16

            Text {
                text: currentItem() ? currentItem().title : "Quickcast"
                font.family: theme.fontFamily
                font.pixelSize: 12
                font.weight: Font.Medium
                color: theme.textSecondary
                elide: Text.ElideRight
                Layout.fillWidth: true
            }

            RowLayout {
                spacing: 10

                // Primary Action Hint
                RowLayout {
                    spacing: 4
                    visible: currentItem() && currentItem().actions && currentItem().actions.length > 0

                    Rectangle {
                        height: 20
                        radius: 4
                        color: theme.bgElevated
                        implicitWidth: enterHint.implicitWidth + 8

                        Text {
                            id: enterHint
                            anchors.centerIn: parent
                            text: "↵"
                            font.family: theme.fontMono
                            font.pixelSize: 11
                            color: theme.accent
                        }
                    }

                    Text {
                        text: (currentItem() && currentItem().actions && currentItem().actions.length > 0) ? currentItem().actions[0].title : ""
                        font.family: theme.fontFamily
                        font.pixelSize: 11
                        color: theme.textSecondary
                    }
                }

                // Actions Hint (Ctrl+K)
                RowLayout {
                    spacing: 4
                    visible: currentItem() && currentItem().actions && currentItem().actions.length > 1

                    Rectangle {
                        height: 20
                        radius: 4
                        color: theme.bgElevated
                        implicitWidth: ctrlKHint.implicitWidth + 8

                        Text {
                            id: ctrlKHint
                            anchors.centerIn: parent
                            text: "Ctrl+K"
                            font.family: theme.fontMono
                            font.pixelSize: 10
                            color: theme.textMuted
                        }
                    }

                    Text {
                        text: "Actions"
                        font.family: theme.fontFamily
                        font.pixelSize: 11
                        color: theme.textSecondary
                    }
                }
            }
        }
    }

    // Action Menu Overlay (Ctrl+K)
    ActionMenu {
        id: actionMenu
        anchors.centerIn: parent
        width: 380
        height: Math.min(320, (actions.length * 40) + 70)
        theme: root.theme
        visible: root.isActionMenuOpen
        actions: (currentItem() && currentItem().actions) ? currentItem().actions : []

        onActionTriggered: function(action) {
            root.isActionMenuOpen = false
            var item = currentItem()
            if (item) {
                root.executeRequested(item.id, action)
            }
        }

        onCloseRequested: {
            root.isActionMenuOpen = false
            searchInput.forceActiveFocus()
        }
    }

    function currentItem() {
        if (root.items && root.selectedIndex >= 0 && root.selectedIndex < root.items.length) {
            return root.items[root.selectedIndex]
        }
        return null
    }

    function selectNextItem() {
        if (resultsList.count > 0) {
            resultsList.currentIndex = (resultsList.currentIndex + 1) % resultsList.count
            root.selectedIndex = resultsList.currentIndex
        }
    }

    function selectPrevItem() {
        if (resultsList.count > 0) {
            resultsList.currentIndex = (resultsList.currentIndex - 1 + resultsList.count) % resultsList.count
            root.selectedIndex = resultsList.currentIndex
        }
    }

    function triggerSelectedItem() {
        var item = currentItem()
        if (item && item.actions && item.actions.length > 0) {
            root.executeRequested(item.id, item.actions[0])
        }
    }

    function toggleActionMenu() {
        var item = currentItem()
        if (item && item.actions && item.actions.length > 0) {
            root.isActionMenuOpen = !root.isActionMenuOpen
        }
    }

    function reset() {
        searchInput.text = ""
        root.selectedIndex = 0
        root.isActionMenuOpen = false
        searchInput.forceActiveFocus()
    }
}
