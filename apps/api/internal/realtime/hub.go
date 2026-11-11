package realtime

import (
	"context"
	"sync"

	"github.com/coder/websocket"
)

type Hub struct {
	mu      sync.Mutex
	clients map[*websocket.Conn]struct{}
}

func NewHub() *Hub {
	return &Hub{clients: make(map[*websocket.Conn]struct{})}
}

func (h *Hub) Add(client *websocket.Conn) {
	h.mu.Lock()
	h.clients[client] = struct{}{}
	h.mu.Unlock()
}

func (h *Hub) Remove(client *websocket.Conn) {
	h.mu.Lock()
	delete(h.clients, client)
	h.mu.Unlock()
}

func (h *Hub) Broadcast(ctx context.Context, payload []byte) {
	h.mu.Lock()
	clients := make([]*websocket.Conn, 0, len(h.clients))
	for client := range h.clients {
		clients = append(clients, client)
	}
	h.mu.Unlock()

	for _, client := range clients {
		if err := client.Write(ctx, websocket.MessageText, payload); err != nil {
			h.Remove(client)
			_ = client.Close(websocket.StatusGoingAway, "write failed")
		}
	}
}
