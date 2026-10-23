package projects

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/coderianx/avenor/internal/store"
	"github.com/go-chi/chi/v5"
)

func DeleteProject(
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

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)

	if err != nil || id <= 0 {
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "invalid project id",
		})

		return
	}

	deleted, err := store.DeleteProject(r.Context(), id)

	if err != nil {
		log.Printf("projects: delete project: %v", err)

		w.WriteHeader(http.StatusInternalServerError)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "failed to delete project",
		})

		return
	}

	if !deleted {
		w.WriteHeader(http.StatusNotFound)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "project not found",
		})

		return
	}

	w.WriteHeader(http.StatusNoContent)
}
