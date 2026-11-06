package ipc

import "github.com/emil/quickcast/internal/search"

// Request is sent from client (CLI or Quickshell) to daemon.
type Request struct {
	ID     int64          `json:"id,omitempty"`
	Action string         `json:"action"` // "search", "execute", "toggle", "show", "hide", "ping"
	Params *RequestParams `json:"params,omitempty"`
}

type RequestParams struct {
	Query  string         `json:"query,omitempty"`
	ItemID string         `json:"item_id,omitempty"`
	Action *search.Action `json:"action,omitempty"`
}

// Response is sent from daemon back to the requester.
type Response struct {
	ID      int64       `json:"id,omitempty"`
	Success bool        `json:"success"`
	Error   string      `json:"error,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// Event is broadcast from daemon to all connected clients (e.g. Quickshell UI).
type Event struct {
	Event string      `json:"event"` // "toggle", "show", "hide", "items_updated"
	Data  interface{} `json:"data,omitempty"`
}
