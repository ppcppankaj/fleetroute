package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"gpsgo/services/m16-activity-log/internal/middleware"
	"gpsgo/services/m16-activity-log/internal/repository"
	"gpsgo/services/m16-activity-log/internal/service"
)

type Handler struct {
	service *service.Service
}

func New(s *service.Service) *Handler { return &Handler{service: s} }

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Use(middleware.Auth)
	r.Get("/", h.listEvents)
}

func (h *Handler) listEvents(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	limit := 100
	if l, err := strconv.Atoi(r.URL.Query().Get("limit")); err == nil && l > 0 {
		limit = l
	}
	events, err := h.service.ListEvents(r.Context(), tenantID, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if events == nil { events = []repository.ActivityEvent{} }
	writeJSON(w, http.StatusOK, events)
}

func writeJSON(w http.ResponseWriter, code int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(body)
}

func writeError(w http.ResponseWriter, code int, message string) {
	writeJSON(w, code, map[string]any{"error": message})
}
