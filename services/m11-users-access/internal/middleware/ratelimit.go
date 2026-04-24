package middleware

import "net/http"

func TenantRateLimit(next http.Handler) http.Handler {
	return next
}
