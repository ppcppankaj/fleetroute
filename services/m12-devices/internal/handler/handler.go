package handler

import (
	"encoding/json"
	"net/http"
	
	"github.com/go-chi/chi/v5"
	"gpsgo/services/m12-devices/internal/service"
	"gpsgo/services/m12-devices/internal/repository"
	"gpsgo/services/m12-devices/internal/middleware"
)

type Handler struct {
	service *service.Service
}

func New(service *service.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Use(middleware.Auth)
	r.Get("/", h.listDevices)
	r.Post("/", h.createDevice)
	r.Get("/{id}", h.getDevice)
	r.Delete("/{id}", h.deleteDevice)
}

func (h *Handler) listDevices(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	ds, err := h.service.ListDevices(r.Context(), tenantID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if ds == nil {
		ds = []repository.Device{}
	}
	writeJSON(w, http.StatusOK, ds)
}

func (h *Handler) getDevice(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id := chi.URLParam(r, "id")
	d, err := h.service.GetDevice(r.Context(), id, tenantID)
	if err != nil {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	writeJSON(w, http.StatusOK, d)
}

func (h *Handler) createDevice(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	var d repository.Device
	if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	d.TenantID = &tenantID
	created, err := h.service.CreateDevice(r.Context(), d)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (h *Handler) deleteDevice(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id := chi.URLParam(r, "id")
	if err := h.service.DeleteDevice(r.Context(), id, tenantID); err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func writeJSON(w http.ResponseWriter, code int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(body)
}

func writeError(w http.ResponseWriter, code int, message string) {
	writeJSON(w, code, map[string]any{"error": message})
}
