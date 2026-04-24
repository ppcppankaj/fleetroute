package response

import (
	"encoding/json"
	"net/http"
)

func JSON(w http.ResponseWriter, code int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(payload)
}

func Error(w http.ResponseWriter, code int, errCode string, extras map[string]any) {
	body := map[string]any{"error": errCode}
	for k, v := range extras {
		body[k] = v
	}
	JSON(w, code, body)
}
