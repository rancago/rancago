package websocket

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/rancago/framework/internal/ports/driven"
)

type HubAdapter struct {
	Register     chan *Client
	Unregister   chan *Client
	BroadcastCh  chan []byte

	mu       sync.RWMutex
	clients  map[*Client]bool
	channels map[string]map[*Client]bool
	direct   map[string]map[*Client]bool
	cache    driven.CachePort
}

type Client struct {
	ID     string
	UserID string
	Send   chan []byte
	hub    *HubAdapter
}

func NewHubAdapter(cache driven.CachePort) *HubAdapter {
	return &HubAdapter{
		Register:    make(chan *Client),
		Unregister:  make(chan *Client),
		BroadcastCh: make(chan []byte, 256),
		clients:     make(map[*Client]bool),
		channels:    make(map[string]map[*Client]bool),
		direct:      make(map[string]map[*Client]bool),
		cache:       cache,
	}
}

func (h *HubAdapter) Run() {
	for {
		select {
		case client := <-h.Register:
			h.mu.Lock()
			h.clients[client] = true
			if client.UserID != "" {
				if _, ok := h.direct[client.UserID]; !ok {
					h.direct[client.UserID] = make(map[*Client]bool)
				}
				h.direct[client.UserID][client] = true
			}
			h.mu.Unlock()
		case client := <-h.Unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.Send)
				if client.UserID != "" {
					if u, ok := h.direct[client.UserID]; ok {
						delete(u, client)
						if len(u) == 0 {
							delete(h.direct, client.UserID)
						}
					}
				}
			}
			h.mu.Unlock()
		case message := <-h.BroadcastCh:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(h.clients, client)
				}
			}
			h.mu.RUnlock()
		}
	}
}

func (h *HubAdapter) Broadcast(raw []byte) {
	select {
	case h.BroadcastCh <- raw:
	default:
		log.Printf("[rancago][ws] broadcast channel full, dropping message")
	}
}

func (h *HubAdapter) SendDirect(userID string, raw []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	clients, ok := h.direct[userID]
	if !ok {
		return
	}
	for c := range clients {
		select {
		case c.Send <- raw:
		default:
			close(c.Send)
			delete(clients, c)
		}
	}
}

func (h *HubAdapter) PublishChannel(channel string, raw []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	clients, ok := h.channels[channel]
	if !ok {
		return
	}
	for c := range clients {
		select {
		case c.Send <- raw:
		default:
			close(c.Send)
			delete(clients, c)
		}
	}
}

func (h *HubAdapter) StartRedisListener() {
	if h.cache == nil {
		return
	}
	go func() {
		_ = h.cache.Subscribe(nil, "ws:broadcast", func(msg []byte) {
			h.Broadcast(msg)
		})
	}()
}

func (h *HubAdapter) Handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-WS-Upgrade", "websocket-not-available-stub")
	msg := driven.WebSocketMessage{
		Type:      "ws:info",
		Channel:   "system",
		Payload:   map[string]interface{}{"note": "SSE/WebSocket stub handler. Extend with gorilla/websocket for real socket support."},
		Timestamp: 0,
	}
	raw, _ := json.Marshal(msg)
	_, _ = w.Write(raw)
}
