package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"gpsgo/services/m13-security/internal/middleware"
	"gpsgo/services/m13-security/internal/repository"
	"gpsgo/services/m13-security/internal/service"
)

type Handler struct {
	service *service.Service
}

func New(s *service.Service) *Handler { return &Handler{service: s} }

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Use(middleware.Auth)
	r.Get("/audit-logs", h.listAuditLogs)
	r.Get("/incidents", h.listIncidents)
}

func (h *Handler) listAuditLogs(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	logs, err := h.service.ListAuditLogs(r.Context(), tenantID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if logs == nil { logs = []repository.AuditLog{} }
	writeJSON(w, http.StatusOK, logs)
}

func (h *Handler) listIncidents(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	incs, err := h.service.ListIncidents(r.Context(), tenantID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if incs == nil { incs = []repository.SecurityIncident{} }
	writeJSON(w, http.StatusOK, incs)
}

func writeJSON(w http.ResponseWriter, code int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(body)
}

func writeError(w http.ResponseWriter, code int, message string) {
	writeJSON(w, code, map[string]any{"error": message})
}

var _ = strconv.Itoa // avoid unused import
