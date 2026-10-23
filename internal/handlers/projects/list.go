package projects

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/coderianx/avenor/internal/store"
)

func ListProjects(
	w http.ResponseWriter,
	r *http.Request,
) {
	w.Header().Set("Content-Type", "application/json")

	if store.DB == nil {
		w.WriteHeader(http.StatusInternalServerError)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "database not connected",
		})

		return
	}

	projects, err := store.ListProjects(r.Context())

	if err != nil {
		log.Printf("projects: list projects: %v", err)

		w.WriteHeader(http.StatusInternalServerError)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "failed to list projects",
		})

		return
	}

	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(projects)
}
