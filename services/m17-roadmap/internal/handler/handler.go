package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"gpsgo/services/m17-roadmap/internal/middleware"
	"gpsgo/services/m17-roadmap/internal/repository"
	"gpsgo/services/m17-roadmap/internal/service"
)

type Handler struct {
	service *service.Service
}

func New(s *service.Service) *Handler { return &Handler{service: s} }

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Use(middleware.Auth)
	r.Get("/", h.listFeatures)
	r.Post("/", h.createFeature)
	r.Post("/{id}/vote", h.castVote)
}

func (h *Handler) listFeatures(w http.ResponseWriter, r *http.Request) {
	fs, err := h.service.ListFeatures(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if fs == nil { fs = []repository.Feature{} }
	writeJSON(w, http.StatusOK, fs)
}

func (h *Handler) createFeature(w http.ResponseWriter, r *http.Request) {
	var f repository.Feature
	if err := json.NewDecoder(r.Body).Decode(&f); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	created, err := h.service.CreateFeature(r.Context(), f)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (h *Handler) castVote(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	userID := middleware.GetUserID(r.Context())
	featureID := chi.URLParam(r, "id")
	if err := h.service.CastVote(r.Context(), featureID, tenantID, userID); err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "voted"})
}

func writeJSON(w http.ResponseWriter, code int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(body)
}

func writeError(w http.ResponseWriter, code int, message string) {
	writeJSON(w, code, map[string]any{"error": message})
}
