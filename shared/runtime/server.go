package runtime

import (
	"net/http"

	"gpsgo/shared/response"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func NewRouter(serviceName string) *chi.Mux {
	r := chi.NewRouter()
	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		response.JSON(w, http.StatusOK, map[string]any{"status": "ok", "service": serviceName})
	})
	r.Handle("/metrics", promhttp.Handler())
	return r
}
