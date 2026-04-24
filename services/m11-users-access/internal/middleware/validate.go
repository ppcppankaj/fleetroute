package middleware

import "net/http"

func Validate(next http.Handler) http.Handler {
	return next
}
