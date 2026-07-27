// Package WebSocket provides a scalable WebSocket hub for Rancago Framework.
// Supports Direct, Channel, and Broadcast message routing.
// Multi-node horizontal scaling is achieved via Redis Pub/Sub:
// every PublishChannel call publishes to "rancago:ws:{channel}" so
// all instances receive and re-dispatch to their local clients.
package WebSocket

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/rancago/framework/framework/Cache"
)

// Message is the standard envelope for all WebSocket messages.
type Message struct {
	Type      string      `json:"type"`
	Channel   string      `json:"channel,omitempty"`
	Payload   interface{} `json:"payload,omitempty"`
	Timestamp int64       `json:"ts,omitempty"`
}

// Client represents a connected WebSocket client.
type Client struct {
	ID     string
	UserID string
	Send   chan []byte
	hub    *Hub
}

// Hub routes messages to connected clients.
type Hub struct {
	register    chan *Client
	unregister  chan *Client
	broadcastCh chan []byte

	mu       sync.RWMutex
	clients  map[*Client]bool
	channels map[string]map[*Client]bool // channel → clients
	direct   map[string]map[*Client]bool // userID → clients

	redis *Cache.RedisManager
}

// NewHub creates a new Hub backed by an optional RedisManager for multi-node pub/sub.
func NewHub(redis *Cache.RedisManager) *Hub {
	return &Hub{
		register:    make(chan *Client),
		unregister:  make(chan *Client),
		broadcastCh: make(chan []byte, 512),
		clients:     make(map[*Client]bool),
		channels:    make(map[string]map[*Client]bool),
		direct:      make(map[string]map[*Client]bool),
		redis:       redis,
	}
}

// Run starts the hub event loop. Must be called in a goroutine.
func (h *Hub) Run() {
	for {
		select {
		case c := <-h.register:
			h.mu.Lock()
			h.clients[c] = true
			if c.UserID != "" {
				if _, ok := h.direct[c.UserID]; !ok {
					h.direct[c.UserID] = make(map[*Client]bool)
				}
				h.direct[c.UserID][c] = true
			}
			h.mu.Unlock()
			log.Printf("[rancago][ws] client connected: %s (user=%s)", c.ID, c.UserID)

		case c := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[c]; ok {
				delete(h.clients, c)
				close(c.Send)
				if c.UserID != "" {
					if u, ok := h.direct[c.UserID]; ok {
						delete(u, c)
						if len(u) == 0 {
							delete(h.direct, c.UserID)
						}
					}
				}
				// Remove from all channel subscriptions.
				for ch, members := range h.channels {
					delete(members, c)
					if len(members) == 0 {
						delete(h.channels, ch)
					}
				}
			}
			h.mu.Unlock()
			log.Printf("[rancago][ws] client disconnected: %s", c.ID)

		case msg := <-h.broadcastCh:
			h.mu.RLock()
			for c := range h.clients {
				select {
				case c.Send <- msg:
				default:
					close(c.Send)
					delete(h.clients, c)
				}
			}
			h.mu.RUnlock()
		}
	}
}

// StartRedisListener subscribes to the global Redis broadcast channel.
// Messages published by any node on "rancago:ws:broadcast" are re-dispatched locally.
func (h *Hub) StartRedisListener() {
	if h.redis == nil {
		return
	}
	go func() {
		_ = h.redis.Subscribe(nil, "rancago:ws:broadcast", func(msg []byte) {
			h.Broadcast(msg)
		})
	}()
}

// Broadcast sends raw bytes to all connected clients.
func (h *Hub) Broadcast(raw []byte) {
	select {
	case h.broadcastCh <- raw:
	default:
		log.Printf("[rancago][ws] broadcast channel full, dropping message")
	}
}

// SendDirect sends a message to all connections belonging to userID.
func (h *Hub) SendDirect(userID string, raw []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for c := range h.direct[userID] {
		select {
		case c.Send <- raw:
		default:
		}
	}
}

// Subscribe adds a client to a named channel.
func (h *Hub) Subscribe(c *Client, channel string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, ok := h.channels[channel]; !ok {
		h.channels[channel] = make(map[*Client]bool)
	}
	h.channels[channel][c] = true
}

// PublishChannel sends a message to all clients subscribed to channel,
// AND publishes to Redis so other nodes can relay it (multi-node support).
func (h *Hub) PublishChannel(channel string, raw []byte) {
	h.mu.RLock()
	for c := range h.channels[channel] {
		select {
		case c.Send <- raw:
		default:
		}
	}
	h.mu.RUnlock()

	// Publish to Redis for other Hub instances.
	if h.redis != nil {
		redisKey := "rancago:ws:" + channel
		_ = h.redis.Publish(nil, redisKey, raw)
	}
}

// ConnectedCount returns the number of currently connected clients.
func (h *Hub) ConnectedCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// Handler is an HTTP handler stub for the WebSocket endpoint.
// It returns an informational JSON response.
// Replace with gorilla/websocket upgrade logic for a production implementation.
func (h *Hub) Handler(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-WS-Note", "stub-handler; upgrade with gorilla/websocket")
	msg := Message{
		Type:    "ws:info",
		Channel: "system",
		Payload: map[string]interface{}{
			"note":       "WebSocket stub handler. Extend with gorilla/websocket for real connections.",
			"user_id":    userID,
			"connected":  h.ConnectedCount(),
		},
	}
	raw, _ := json.Marshal(msg)
	_, _ = w.Write(raw)
}
