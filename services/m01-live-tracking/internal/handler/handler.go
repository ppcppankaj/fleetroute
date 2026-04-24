package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	
	"github.com/go-chi/chi/v5"
	"gpsgo/services/m01-live-tracking/internal/service"
	"gpsgo/services/m01-live-tracking/internal/websocket"
	"gpsgo/services/m01-live-tracking/internal/repository"
	"gpsgo/services/m01-live-tracking/internal/middleware"
)

type Handler struct {
	service *service.Service
	hub     *websocket.Hub
}

func New(service *service.Service, hub *websocket.Hub) *Handler {
	return &Handler{service: service, hub: hub}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Use(middleware.Auth)
	
	// REST
	r.Get("/vehicles/{id}/breadcrumbs", h.getBreadcrumbs)
	
	// WebSocket
	r.Get("/ws", h.serveWS)
}

func (h *Handler) getBreadcrumbs(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	vehicleID := chi.URLParam(r, "id")
	
	limitStr := r.URL.Query().Get("limit")
	limit := 100
	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
		limit = l
	}

	bs, err := h.service.GetBreadcrumbs(r.Context(), vehicleID, tenantID, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if bs == nil {
		bs = []repository.Breadcrumb{}
	}
	writeJSON(w, http.StatusOK, bs)
}

func (h *Handler) serveWS(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	h.hub.ServeWS(w, r, tenantID)
}

func writeJSON(w http.ResponseWriter, code int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(body)
}

func writeError(w http.ResponseWriter, code int, message string) {
	writeJSON(w, code, map[string]any{"error": message})
}
