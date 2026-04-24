package middleware

import "net/http"

func TrustGatewayHeaders(next http.Handler) http.Handler {
	return next
}
