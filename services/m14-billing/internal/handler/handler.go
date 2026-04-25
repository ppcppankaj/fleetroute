package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"gpsgo/services/m14-billing/internal/middleware"
	"gpsgo/services/m14-billing/internal/repository"
	"gpsgo/services/m14-billing/internal/service"
)

type Handler struct {
	service *service.Service
}

func New(s *service.Service) *Handler { return &Handler{service: s} }

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Use(middleware.Auth)
	r.Get("/subscription", h.getSubscription)
	r.Get("/invoices", h.listInvoices)
	r.Post("/invoices", h.createInvoice)
}

func (h *Handler) getSubscription(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	sub, err := h.service.GetSubscription(r.Context(), tenantID)
	if err != nil {
		writeError(w, http.StatusNotFound, "no subscription found")
		return
	}
	writeJSON(w, http.StatusOK, sub)
}

func (h *Handler) listInvoices(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	invs, err := h.service.ListInvoices(r.Context(), tenantID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if invs == nil { invs = []repository.Invoice{} }
	writeJSON(w, http.StatusOK, invs)
}

func (h *Handler) createInvoice(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	var inv repository.Invoice
	if err := json.NewDecoder(r.Body).Decode(&inv); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	inv.TenantID = tenantID
	created, err := h.service.CreateInvoice(r.Context(), inv)
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
