import QtQuick
import Quickshell
import Quickshell.Wayland
import Quickshell.Io

ShellRoot {
    id: root

    property bool isVisible: false

    PanelWindow {
        id: window
        visible: root.isVisible

        WlrLayershell.layer: WlrLayershell.Overlay
        WlrLayershell.keyboardFocus: root.isVisible ? WlrLayershell.KeyboardFocusExclusive : WlrLayershell.KeyboardFocusNone
        WlrLayershell.namespace: "quickcast"

        // Position centered on screen
        anchors {
            // When anchors are omitted in Quickshell, it centers automatically
        }

        color: "transparent"
        width: launcher.width
        height: launcher.height

        Launcher {
            id: launcher

            onQueryChanged: function(text) {
                sendIpcRequest({
                    action: "search",
                    params: { query: text }
                })
            }

            onExecuteRequested: function(itemId, action) {
                sendIpcRequest({
                    action: "execute",
                    params: {
                        item_id: itemId,
                        action: action
                    }
                })
                root.isVisible = false
            }

            onCloseRequested: {
                root.isVisible = false
            }
        }
    }

    // Unix Domain Socket connecting to Go backend daemon
    Socket {
        id: daemonSocket
        path: (Quickshell.env("XDG_RUNTIME_DIR") || "/tmp") + "/quickcast.sock"
        connected: true

        parser: SplitParser {
            splitMarker: "\n"
            onRead: function(data) {
                if (!data || data.length === 0) return
                try {
                    var msg = JSON.parse(data)
                    handleIpcMessage(msg)
                } catch (e) {
                    console.error("Quickcast IPC parse error:", e, data)
                }
            }
        }
    }

    function handleIpcMessage(msg) {
        // Event broadcast from daemon (e.g. CLI triggered "quickcast toggle")
        if (msg.event) {
            if (msg.event === "toggle") {
                root.isVisible = !root.isVisible
                if (root.isVisible) {
                    launcher.reset()
                    // Initial search with empty query to load recent/suggested items
                    sendIpcRequest({ action: "search", params: { query: "" } })
                }
            } else if (msg.event === "show") {
                root.isVisible = true
                launcher.reset()
                sendIpcRequest({ action: "search", params: { query: "" } })
            } else if (msg.event === "hide") {
                root.isVisible = false
            }
            return
        }

        // Response to search / execute request
        if (msg.data !== undefined) {
            if (Array.isArray(msg.data)) {
                launcher.items = msg.data
                launcher.selectedIndex = 0
            }
        }
    }

    function sendIpcRequest(req) {
        if (!daemonSocket.connected) {
            daemonSocket.connected = true
        }
        var raw = JSON.stringify(req) + "\n"
        daemonSocket.write(raw)
    }
}
