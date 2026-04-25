package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"gpsgo/services/m05-reports/internal/middleware"
	"gpsgo/services/m05-reports/internal/repository"
	"gpsgo/services/m05-reports/internal/service"
)

type Handler struct {
	service *service.Service
}

func New(s *service.Service) *Handler { return &Handler{service: s} }

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Use(middleware.Auth)
	r.Get("/definitions", h.listDefinitions)
	r.Post("/definitions", h.createDefinition)
	r.Get("/runs", h.listRuns)
	r.Post("/definitions/{id}/run", h.runReport)
}

func (h *Handler) listDefinitions(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	ds, err := h.service.ListDefinitions(r.Context(), tenantID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if ds == nil { ds = []repository.ReportDefinition{} }
	writeJSON(w, http.StatusOK, ds)
}

func (h *Handler) createDefinition(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	var d repository.ReportDefinition
	if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	d.TenantID = tenantID
	created, err := h.service.CreateDefinition(r.Context(), d)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (h *Handler) runReport(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	defID := chi.URLParam(r, "id")
	run, err := h.service.RunReport(r.Context(), defID, tenantID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	writeJSON(w, http.StatusAccepted, run)
}

func (h *Handler) listRuns(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	runs, err := h.service.ListRuns(r.Context(), tenantID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if runs == nil { runs = []repository.ReportRun{} }
	writeJSON(w, http.StatusOK, runs)
}

func writeJSON(w http.ResponseWriter, code int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(body)
}

func writeError(w http.ResponseWriter, code int, message string) {
	writeJSON(w, code, map[string]any{"error": message})
}
