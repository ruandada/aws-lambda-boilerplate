package app

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

func NewMux() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, _ *http.Request) {
		respondJSON(w, http.StatusOK, map[string]any{
			"message": "Hello World",
		})
	})

	mux.HandleFunc("GET /api/greet/{name}", func(w http.ResponseWriter, r *http.Request) {
		name := r.PathValue("name")
		from := strings.TrimSpace(r.URL.Query().Get("from"))
		if from == "" {
			from = "starter"
		}

		respondJSON(w, http.StatusOK, map[string]any{
			"message": fmt.Sprintf("Hello, %s!", name),
			"from":    from,
		})
	})

	return mux
}

func StartLocalServer(port int) error {
	mux := NewMux()
	addr := fmt.Sprintf(":%d", port)
	log.Printf("Local server is running on http://localhost:%d", port)
	return http.ListenAndServe(addr, mux)
}

func respondJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
