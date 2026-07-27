package http

import (
	"encoding/json"
	"net/http"

	"github.com/rancago/framework/internal/domain/entities"
	"github.com/rancago/framework/internal/domain/valueobjects"
	"github.com/rancago/framework/internal/ports/driving"
)

type HealthHandler struct {
	appName string
}

func NewHealthHandler(appName string) *HealthHandler {
	return &HealthHandler{appName: appName}
}

func (h *HealthHandler) Welcome(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Welcome to Rancago Framework 🚀",
		"version": "1.0.0",
		"engine":  h.appName,
	})
}

func (h *HealthHandler) Health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":  "healthy",
		"service": "rancago-api",
	})
}

func writeJSON(w http.ResponseWriter, code int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

type NotificationHandler struct {
	uc driving.NotificationUseCase
}

func NewNotificationHandler(uc driving.NotificationUseCase) *NotificationHandler {
	return &NotificationHandler{uc: uc}
}

func (h *NotificationHandler) RegisterRoutes(mux *http.ServeMux, prefix string) {
	mux.HandleFunc(prefix+"/send", h.handleSend)
	mux.HandleFunc(prefix+"/broadcast", h.handleBroadcast)
	mux.HandleFunc(prefix+"/list", h.handleList)
	mux.HandleFunc(prefix+"/count", h.handleCount)
	mux.HandleFunc(prefix+"/read", h.handleMarkRead)
}

type sendRequest struct {
	UserID  string            `json:"user_id"`
	Title   string            `json:"title"`
	Body    string            `json:"body"`
	Channel string            `json:"channel"`
	Data    map[string]string `json:"data"`
}

func (h *NotificationHandler) handleSend(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, errResp("method not allowed"))
		return
	}
	var req sendRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errResp("invalid body: "+err.Error()))
		return
	}
	if req.Channel == "" {
		req.Channel = "database"
	}
	ch := parseChannel(req.Channel)
	n, err := h.uc.Send(r.Context(), req.UserID, req.Title, req.Body, ch, req.Data)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errResp(err.Error()))
		return
	}
	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"data": n,
	})
}

type broadcastRequest struct {
	Title string            `json:"title"`
	Body  string            `json:"body"`
	Data  map[string]string `json:"data"`
}

func (h *NotificationHandler) handleBroadcast(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, errResp("method not allowed"))
		return
	}
	var req broadcastRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errResp("invalid body: "+err.Error()))
		return
	}
	if err := h.uc.Broadcast(r.Context(), req.Title, req.Body, req.Data); err != nil {
		writeJSON(w, http.StatusBadRequest, errResp(err.Error()))
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":  "broadcasted",
		"message": "notification sent to all connected clients",
	})
}

func (h *NotificationHandler) handleList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, errResp("method not allowed"))
		return
	}
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		writeJSON(w, http.StatusBadRequest, errResp("missing user_id query parameter"))
		return
	}
	page := atoiDefault(r.URL.Query().Get("page"), 1)
	perPage := atoiDefault(r.URL.Query().Get("per_page"), 25)
	items, meta, err := h.uc.ListUserNotifications(r.Context(), userID, page, perPage)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errResp(err.Error()))
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"data": items,
		"meta": meta,
	})
}

func (h *NotificationHandler) handleCount(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, errResp("method not allowed"))
		return
	}
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		writeJSON(w, http.StatusBadRequest, errResp("missing user_id query parameter"))
		return
	}
	cnt, err := h.uc.GetUnreadCount(r.Context(), userID)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errResp(err.Error()))
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"user_id":      userID,
		"unread_count": cnt,
	})
}

type markReadRequest struct {
	ID     string `json:"id"`
	UserID string `json:"user_id"`
}

func (h *NotificationHandler) handleMarkRead(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, errResp("method not allowed"))
		return
	}
	var req markReadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errResp("invalid body: "+err.Error()))
		return
	}
	id := parseID(req.ID)
	if err := h.uc.MarkRead(r.Context(), id, req.UserID); err != nil {
		writeJSON(w, http.StatusBadRequest, errResp(err.Error()))
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status": "read",
		"id":     req.ID,
	})
}

func errResp(msg string) map[string]interface{} {
	return map[string]interface{}{"error": msg}
}

func parseChannel(s string) entities.NotificationChannel {
	switch s {
	case "email":
		return entities.ChannelEmail
	case "push":
		return entities.ChannelPush
	case "sms":
		return entities.ChannelSMS
	case "broadcast":
		return entities.ChannelBroadcast
	default:
		return entities.ChannelDatabase
	}
}

func parseID(s string) valueobjects.ID {
	return valueobjects.NewIDStr(s)
}

func atoiDefault(s string, def int) int {
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
