package api

import (
	"log/slog"
	"net/http"
	"sync"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type WebSocketManager struct {
	sync.RWMutex
	subscriptions map[string]map[*websocket.Conn]struct{}
}

func NewWebSocketManager() *WebSocketManager {
	return &WebSocketManager{
		subscriptions: make(map[string]map[*websocket.Conn]struct{}),
	}
}

func (m *WebSocketManager) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	lotID := chi.URLParam(r, "id")
	if lotID == "" {
		http.Error(w, "missing lot id", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("websocket upgrade failed", "error", err)
		return
	}
	defer func() {
		if err := conn.Close(); err != nil {
			slog.Error("websocket close error", "error", err)
		}
	}()

	m.Lock()
	if m.subscriptions[lotID] == nil {
		m.subscriptions[lotID] = make(map[*websocket.Conn]struct{})
	}
	m.subscriptions[lotID][conn] = struct{}{}
	m.Unlock()

	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			break
		}
	}

	m.Lock()
	delete(m.subscriptions[lotID], conn)
	if len(m.subscriptions[lotID]) == 0 {
		delete(m.subscriptions, lotID)
	}
	m.Unlock()
}

func (m *WebSocketManager) BroadcastToLot(lotID string, message interface{}) {
	m.RLock()
	defer m.RUnlock()
	conns, ok := m.subscriptions[lotID]
	if !ok {
		return
	}
	for conn := range conns {
		if err := conn.WriteJSON(message); err != nil {
			slog.Error("websocket write error", "lot_id", lotID, "error", err)
			if err := conn.Close(); err != nil {
				slog.Error("failed to close websocket connection", "lot_id", lotID, "error", err)
			}
		}
	}
}
