package handler

import (
	"encoding/json"
	"net/http"
	
	"github.com/go-chi/chi/v5"
	"gpsgo/services/m07-drivers/internal/service"
	"gpsgo/services/m07-drivers/internal/repository"
	"gpsgo/services/m07-drivers/internal/middleware"
)

type Handler struct {
	service *service.Service
}

func New(service *service.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Use(middleware.Auth)
	r.Get("/", h.listDrivers)
	r.Post("/", h.createDriver)
	r.Get("/{id}", h.getDriver)
	r.Delete("/{id}", h.deleteDriver)
}

func (h *Handler) listDrivers(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	ds, err := h.service.ListDrivers(r.Context(), tenantID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if ds == nil {
		ds = []repository.Driver{}
	}
	writeJSON(w, http.StatusOK, ds)
}

func (h *Handler) getDriver(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id := chi.URLParam(r, "id")
	d, err := h.service.GetDriver(r.Context(), id, tenantID)
	if err != nil {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	writeJSON(w, http.StatusOK, d)
}

func (h *Handler) createDriver(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	var d repository.Driver
	if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	d.TenantID = tenantID
	created, err := h.service.CreateDriver(r.Context(), d)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (h *Handler) deleteDriver(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id := chi.URLParam(r, "id")
	if err := h.service.DeleteDriver(r.Context(), id, tenantID); err != nil {
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
