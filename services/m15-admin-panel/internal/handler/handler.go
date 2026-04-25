package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"gpsgo/services/m15-admin-panel/internal/middleware"
	"gpsgo/services/m15-admin-panel/internal/repository"
	"gpsgo/services/m15-admin-panel/internal/service"
)

type Handler struct {
	service   *service.Service
	jwtSecret string
}

func New(s *service.Service, jwtSecret string) *Handler {
	return &Handler{service: s, jwtSecret: jwtSecret}
}

func (h *Handler) RegisterPublicRoutes(r chi.Router) {
	r.Post("/auth/login", h.login)
}

func (h *Handler) RegisterProtectedRoutes(r chi.Router) {
	r.Use(middleware.AdminAuth(h.jwtSecret))
	r.Get("/tickets", h.listTickets)
	r.Post("/tickets", h.createTicket)
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	result, err := h.service.Login(r.Context(), body.Email, body.Password)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) listTickets(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	tickets, err := h.service.ListTickets(r.Context(), status)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if tickets == nil { tickets = []repository.SupportTicket{} }
	writeJSON(w, http.StatusOK, tickets)
}

func (h *Handler) createTicket(w http.ResponseWriter, r *http.Request) {
	var t repository.SupportTicket
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	created, err := h.service.CreateTicket(r.Context(), t)
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
