package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"gpsgo/services/m09-fuel/internal/middleware"
	"gpsgo/services/m09-fuel/internal/repository"
	"gpsgo/services/m09-fuel/internal/service"
)

type Handler struct {
	service *service.Service
}

func New(s *service.Service) *Handler {
	return &Handler{service: s}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Use(middleware.Auth)
	r.Get("/", h.listFuelLogs)
	r.Post("/", h.createFuelLog)
}

func (h *Handler) listFuelLogs(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	logs, err := h.service.ListFuelLogs(r.Context(), tenantID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if logs == nil {
		logs = []repository.FuelLog{}
	}
	writeJSON(w, http.StatusOK, logs)
}

func (h *Handler) createFuelLog(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	var f repository.FuelLog
	if err := json.NewDecoder(r.Body).Decode(&f); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	f.TenantID = tenantID
	created, err := h.service.CreateFuelLog(r.Context(), f)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func writeJSON(w http.ResponseWriter, code int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(body)
}

func writeError(w http.ResponseWriter, code int, message string) {
	writeJSON(w, code, map[string]any{"error": message})
}
