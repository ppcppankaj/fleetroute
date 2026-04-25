package handler

import (
	"encoding/json"
	"net/http"
	
	"github.com/go-chi/chi/v5"
	"gpsgo/services/m04-alerts/internal/service"
	"gpsgo/services/m04-alerts/internal/repository"
	"gpsgo/services/m04-alerts/internal/middleware"
)

type Handler struct {
	service *service.Service
}

func New(service *service.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Use(middleware.Auth)
	r.Get("/", h.listAlerts)
	r.Post("/{id}/resolve", h.resolveAlert)
}

func (h *Handler) listAlerts(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	alerts, err := h.service.ListActiveAlerts(r.Context(), tenantID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if alerts == nil {
		alerts = []repository.ActiveAlert{}
	}
	writeJSON(w, http.StatusOK, alerts)
}

func (h *Handler) resolveAlert(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	userID := r.Context().Value(middleware.UserIDKey).(string)
	id := chi.URLParam(r, "id")
	
	err := h.service.ResolveAlert(r.Context(), id, tenantID, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "resolved"})
}

func writeJSON(w http.ResponseWriter, code int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(body)
}

func writeError(w http.ResponseWriter, code int, message string) {
	writeJSON(w, code, map[string]any{"error": message})
}
