package providers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/coderianx/avenor/internal/store"
)

func ListProviders(
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

	providers, err := store.ListProviders(r.Context())

	if err != nil {
		log.Printf("providers: list providers: %v", err)

		w.WriteHeader(http.StatusInternalServerError)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "failed to list providers",
		})

		return
	}

	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(providers)
}
