package ipc

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sync"

	"github.com/emil/quickcast/internal/search"
)

type Handler interface {
	OnSearch(query string) []search.Item
	OnExecute(itemID string, action search.Action) error
}

type Server struct {
	socketPath string
	listener   net.Listener
	handler    Handler

	mu      sync.RWMutex
	clients map[net.Conn]bool
	quit    chan struct{}
}

func GetSocketPath() string {
	runtimeDir := os.Getenv("XDG_RUNTIME_DIR")
	if runtimeDir == "" {
		runtimeDir = os.TempDir()
	}
	return filepath.Join(runtimeDir, "quickcast.sock")
}

func NewServer(handler Handler, customSocket ...string) *Server {
	sock := GetSocketPath()
	if len(customSocket) > 0 && customSocket[0] != "" {
		sock = customSocket[0]
	}

	return &Server{
		socketPath: sock,
		handler:    handler,
		clients:    make(map[net.Conn]bool),
		quit:       make(chan struct{}),
	}
}

func (s *Server) Start() error {
	// Remove stale socket if it exists
	_ = os.Remove(s.socketPath)

	if err := os.MkdirAll(filepath.Dir(s.socketPath), 0755); err != nil {
		return err
	}

	l, err := net.Listen("unix", s.socketPath)
	if err != nil {
		return fmt.Errorf("failed to listen on socket %s: %w", s.socketPath, err)
	}
	s.listener = l

	go s.acceptLoop()
	return nil
}

func (s *Server) Stop() {
	close(s.quit)
	if s.listener != nil {
		_ = s.listener.Close()
	}
	_ = os.Remove(s.socketPath)

	s.mu.Lock()
	for c := range s.clients {
		_ = c.Close()
	}
	s.clients = make(map[net.Conn]bool)
	s.mu.Unlock()
}

func (s *Server) Broadcast(event Event) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, err := json.Marshal(event)
	if err != nil {
		return
	}
	data = append(data, '\n')

	for conn := range s.clients {
		go func(c net.Conn) {
			_, _ = c.Write(data)
		}(conn)
	}
}

func (s *Server) acceptLoop() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-s.quit:
				return
			default:
				continue
			}
		}

		s.mu.Lock()
		s.clients[conn] = true
		s.mu.Unlock()

		go s.handleClient(conn)
	}
}

func (s *Server) handleClient(conn net.Conn) {
	defer func() {
		s.mu.Lock()
		delete(s.clients, conn)
		s.mu.Unlock()
		_ = conn.Close()
	}()

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var req Request
		if err := json.Unmarshal(line, &req); err != nil {
			s.sendResp(conn, Response{ID: req.ID, Success: false, Error: "invalid JSON"})
			continue
		}

		resp := s.processRequest(req)
		s.sendResp(conn, resp)
	}
}

func (s *Server) processRequest(req Request) Response {
	switch req.Action {
	case "ping":
		return Response{ID: req.ID, Success: true, Data: "pong"}

	case "toggle":
		s.Broadcast(Event{Event: "toggle"})
		return Response{ID: req.ID, Success: true}

	case "show":
		s.Broadcast(Event{Event: "show"})
		return Response{ID: req.ID, Success: true}

	case "hide":
		s.Broadcast(Event{Event: "hide"})
		return Response{ID: req.ID, Success: true}

	case "search":
		query := ""
		if req.Params != nil {
			query = req.Params.Query
		}
		items := s.handler.OnSearch(query)
		return Response{ID: req.ID, Success: true, Data: items}

	case "execute":
		if req.Params == nil || req.Params.Action == nil {
			return Response{ID: req.ID, Success: false, Error: "missing action parameter"}
		}
		err := s.handler.OnExecute(req.Params.ItemID, *req.Params.Action)
		if err != nil {
			return Response{ID: req.ID, Success: false, Error: err.Error()}
		}
		// Raycast auto-hides after executing an action unless configured otherwise
		s.Broadcast(Event{Event: "hide"})
		return Response{ID: req.ID, Success: true}

	default:
		return Response{ID: req.ID, Success: false, Error: "unknown action: " + req.Action}
	}
}

func (s *Server) sendResp(conn net.Conn, resp Response) {
	data, err := json.Marshal(resp)
	if err != nil {
		return
	}
	data = append(data, '\n')
	_, _ = conn.Write(data)
}
