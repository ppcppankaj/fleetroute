package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"gpsgo/services/m02-routes-trips/internal/middleware"
	"gpsgo/services/m02-routes-trips/internal/repository"
	"gpsgo/services/m02-routes-trips/internal/service"
)

type Handler struct {
	service *service.Service
}

func New(s *service.Service) *Handler {
	return &Handler{service: s}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Use(middleware.Auth)
	r.Get("/trips", h.listTrips)
	r.Get("/routes", h.listRoutes)
	r.Post("/routes", h.createRoute)
}

func (h *Handler) listTrips(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	trips, err := h.service.ListTrips(r.Context(), tenantID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if trips == nil {
		trips = []repository.Trip{}
	}
	writeJSON(w, http.StatusOK, trips)
}

func (h *Handler) listRoutes(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	routes, err := h.service.ListRoutes(r.Context(), tenantID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if routes == nil {
		routes = []repository.Route{}
	}
	writeJSON(w, http.StatusOK, routes)
}

func (h *Handler) createRoute(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	var ro repository.Route
	if err := json.NewDecoder(r.Body).Decode(&ro); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	ro.TenantID = tenantID
	created, err := h.service.CreateRoute(r.Context(), ro)
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
