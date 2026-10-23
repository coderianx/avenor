package providers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/coderianx/avenor/internal/models"
	"github.com/coderianx/avenor/internal/secrets"
	"github.com/coderianx/avenor/internal/store"
)

func AddNewProvider(
	w http.ResponseWriter,
	r *http.Request,
) {
	w.Header().Set("Content-Type", "application/json")

	if store.DB == nil {
		w.WriteHeader(500)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "database not connected",
		})

		return
	}

	var req models.AddNewProviderReq

	err := json.NewDecoder(r.Body).Decode(&req)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "invalid request",
		})

		return
	}

	req.Name = strings.TrimSpace(req.Name)
	req.BaseURL = strings.TrimSpace(req.BaseURL)
	req.APIKey = strings.TrimSpace(req.APIKey)

	if req.Name == "" {
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "name is required",
		})

		return
	}

	if req.BaseURL == "" {
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "base_url is required",
		})

		return
	}

	if req.APIKey == "" {
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "api_key is required",
		})

		return
	}

	sealed, err := secrets.Seal(req.APIKey)

	if err != nil {
		log.Printf("providers: seal api key: %v", err)

		w.WriteHeader(http.StatusInternalServerError)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "failed to encrypt api key",
		})

		return
	}

	last4 := last4(req.APIKey)

	id, err := store.CreateProvider(
		r.Context(),
		req.Name,
		req.BaseURL,
		last4,
		sealed,
	)

	if err != nil {
		if store.IsUniqueViolation(err) {
			w.WriteHeader(http.StatusConflict)

			json.NewEncoder(w).Encode(map[string]any{
				"error": "a provider with this base_url already exists",
			})

			return
		}

		log.Printf("providers: create provider: %v", err)

		w.WriteHeader(http.StatusInternalServerError)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "failed to create provider",
		})

		return
	}

	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(models.ProviderRes{
		ID:          id,
		Name:        req.Name,
		BaseURL:     req.BaseURL,
		APIKeySet:   true,
		APIKeyLast4: last4,
	})
}

// last4 returns the last four characters of key so the stored hint is enough
// to recognize a key without exposing short keys.
func last4(key string) string {
	if len(key) <= 4 {
		return ""
	}

	return key[len(key)-4:]
}
