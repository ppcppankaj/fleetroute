package handler

import (
	"encoding/json"
	"net/http"
	
	"github.com/go-chi/chi/v5"
	"gpsgo/services/m06-vehicles/internal/service"
	"gpsgo/services/m06-vehicles/internal/repository"
	"gpsgo/services/m06-vehicles/internal/middleware"
)

type Handler struct {
	service *service.Service
}

func New(service *service.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Use(middleware.Auth)
	r.Get("/", h.listVehicles)
	r.Post("/", h.createVehicle)
	r.Get("/{id}", h.getVehicle)
	r.Delete("/{id}", h.deleteVehicle)
}

func (h *Handler) listVehicles(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	vs, err := h.service.ListVehicles(r.Context(), tenantID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if vs == nil {
		vs = []repository.Vehicle{}
	}
	writeJSON(w, http.StatusOK, vs)
}

func (h *Handler) getVehicle(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id := chi.URLParam(r, "id")
	v, err := h.service.GetVehicle(r.Context(), id, tenantID)
	if err != nil {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	writeJSON(w, http.StatusOK, v)
}

func (h *Handler) createVehicle(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	var v repository.Vehicle
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	v.TenantID = tenantID
	created, err := h.service.CreateVehicle(r.Context(), v)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (h *Handler) deleteVehicle(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id := chi.URLParam(r, "id")
	if err := h.service.DeleteVehicle(r.Context(), id, tenantID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
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
