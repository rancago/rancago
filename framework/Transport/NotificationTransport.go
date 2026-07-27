// Package Transport provides multi-API transport adapters for Rancago Framework.
// The same service is exposed via REST, gRPC, and WebSocket without code duplication.
// Business logic lives in app/Services — these adapters only handle format translation.
package Transport

import (
	"encoding/json"
	"net/http"

	"github.com/rancago/framework/app/Contracts"
)

// ---- REST Adapter ----

// NotificationRESTAdapter exposes NotificationService via HTTP JSON.
type NotificationRESTAdapter struct {
	svc Contracts.NotificationService
}

// NewRESTAdapter creates a REST adapter for the notification service.
func NewRESTAdapter(svc Contracts.NotificationService) *NotificationRESTAdapter {
	return &NotificationRESTAdapter{svc: svc}
}

// RegisterRoutes registers the notification REST routes on mux under prefix.
// Prefix example: "/api/v1/notifications".
func (a *NotificationRESTAdapter) RegisterRoutes(mux *http.ServeMux, prefix string) {
	mux.HandleFunc(prefix+"/send", a.handleSend)
	mux.HandleFunc(prefix+"/broadcast", a.handleBroadcast)
	mux.HandleFunc(prefix+"/list", a.handleList)
	mux.HandleFunc(prefix+"/count", a.handleCount)
	mux.HandleFunc(prefix+"/read", a.handleMarkRead)
}

type sendReq struct {
	UserID  string            `json:"user_id"`
	Title   string            `json:"title"`
	Body    string            `json:"body"`
	Channel string            `json:"channel"`
	Data    map[string]string `json:"data"`
}

func (a *NotificationRESTAdapter) handleSend(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, errBody("method not allowed"))
		return
	}
	var req sendReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errBody("invalid body: "+err.Error()))
		return
	}
	if req.Channel == "" {
		req.Channel = "database"
	}
	n := &Contracts.Notification{
		UserID:  req.UserID,
		Title:   req.Title,
		Body:    req.Body,
		Channel: Contracts.NotificationChannel(req.Channel),
		Data:    req.Data,
	}
	created, err := a.svc.Send(r.Context(), n)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errBody(err.Error()))
		return
	}
	writeJSON(w, http.StatusCreated, map[string]interface{}{"data": created})
}

type broadcastReq struct {
	Title string            `json:"title"`
	Body  string            `json:"body"`
	Data  map[string]string `json:"data"`
}

func (a *NotificationRESTAdapter) handleBroadcast(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, errBody("method not allowed"))
		return
	}
	var req broadcastReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errBody("invalid body: "+err.Error()))
		return
	}
	if err := a.svc.Broadcast(r.Context(), req.Title, req.Body, req.Data); err != nil {
		writeJSON(w, http.StatusBadRequest, errBody(err.Error()))
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"status": "broadcasted"})
}

func (a *NotificationRESTAdapter) handleList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, errBody("method not allowed"))
		return
	}
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		writeJSON(w, http.StatusBadRequest, errBody("missing user_id"))
		return
	}
	limit := queryInt(r, "limit", 25)
	offset := queryInt(r, "offset", 0)
	items, meta, err := a.svc.List(r.Context(), userID, limit, offset)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errBody(err.Error()))
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"data": items, "meta": meta})
}

func (a *NotificationRESTAdapter) handleCount(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, errBody("method not allowed"))
		return
	}
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		writeJSON(w, http.StatusBadRequest, errBody("missing user_id"))
		return
	}
	cnt, err := a.svc.GetUnreadCount(r.Context(), userID)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errBody(err.Error()))
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"user_id": userID, "unread": cnt})
}

type markReadReq struct {
	ID     string `json:"id"`
	UserID string `json:"user_id"`
}

func (a *NotificationRESTAdapter) handleMarkRead(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, errBody("method not allowed"))
		return
	}
	var req markReadReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errBody("invalid body: "+err.Error()))
		return
	}
	if err := a.svc.MarkRead(r.Context(), req.ID, req.UserID); err != nil {
		writeJSON(w, http.StatusBadRequest, errBody(err.Error()))
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"status": "read", "id": req.ID})
}

// ---- gRPC Adapter ----

// GRPCServer is a minimal interface satisfied by any gRPC server.
// Replace with *grpc.Server when google.golang.org/grpc is added.
type GRPCServer interface{}

// NotificationGRPCAdapter registers gRPC service handlers for notifications.
type NotificationGRPCAdapter struct {
	svc Contracts.NotificationService
}

// NewGRPCAdapter creates a gRPC adapter for the notification service.
func NewGRPCAdapter(svc Contracts.NotificationService) *NotificationGRPCAdapter {
	return &NotificationGRPCAdapter{svc: svc}
}

// RegisterGRPC registers the notification handlers on the gRPC server.
// Extend this with generated protobuf service registration when grpc is added.
func (a *NotificationGRPCAdapter) RegisterGRPC(_ GRPCServer) {
	// TODO: register pb.RegisterNotificationServiceServer(s, a) once grpc is wired.
}

// ---- WebSocket Action Handler ----

// WSAction describes a WebSocket action and its handler.
type WSAction struct {
	Action  string
	Handler func(ctx interface{}, payload map[string]interface{}) (interface{}, error)
}

// NotificationWSAdapter handles WebSocket action envelopes for notifications.
type NotificationWSAdapter struct {
	svc     Contracts.NotificationService
	actions map[string]WSAction
}

// NewWSAdapter creates a WebSocket action adapter for notifications.
func NewWSAdapter(svc Contracts.NotificationService) *NotificationWSAdapter {
	a := &NotificationWSAdapter{
		svc:     svc,
		actions: make(map[string]WSAction),
	}
	a.registerActions()
	return a
}

func (a *NotificationWSAdapter) registerActions() {
	// Placeholder — in a real implementation these are dispatch via the Hub.
	// action: "notification:send", "notification:list", etc.
}

// Actions returns all registered action names.
func (a *NotificationWSAdapter) Actions() []string {
	names := make([]string, 0, len(a.actions))
	for k := range a.actions {
		names = append(names, k)
	}
	return names
}

// ---- helpers ----

func writeJSON(w http.ResponseWriter, code int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func errBody(msg string) map[string]interface{} { return map[string]interface{}{"error": msg} }

func queryInt(r *http.Request, key string, def int) int {
	s := r.URL.Query().Get(key)
	if s == "" {
		return def
	}
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return def
		}
		n = n*10 + int(c-'0')
	}
	if n <= 0 {
		return def
	}
	return n
}
