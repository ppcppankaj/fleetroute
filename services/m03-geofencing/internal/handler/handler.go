package handler

import (
	"encoding/json"
	"net/http"
	
	"github.com/go-chi/chi/v5"
	"gpsgo/services/m03-geofencing/internal/service"
	"gpsgo/services/m03-geofencing/internal/repository"
	"gpsgo/services/m03-geofencing/internal/middleware"
)

type Handler struct {
	service *service.Service
}

func New(service *service.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Use(middleware.Auth)
	r.Get("/", h.listGeofences)
	r.Post("/", h.createGeofence)
}

func (h *Handler) listGeofences(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	gs, err := h.service.ListGeofences(r.Context(), tenantID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if gs == nil {
		gs = []repository.Geofence{}
	}
	writeJSON(w, http.StatusOK, gs)
}

func (h *Handler) createGeofence(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	var g repository.Geofence
	if err := json.NewDecoder(r.Body).Decode(&g); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	g.TenantID = tenantID
	created, err := h.service.CreateGeofence(r.Context(), g)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
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
