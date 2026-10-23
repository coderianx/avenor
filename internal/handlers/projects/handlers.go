package projects

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/coderianx/avenor/internal/models"
	"github.com/coderianx/avenor/internal/store"
)

func AddNewProject(
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

	var req models.AddNewProjectReq

	err := json.NewDecoder(r.Body).Decode(&req)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "invalid request",
		})

		return
	}

	req.Name = strings.TrimSpace(req.Name)
	req.RootPath = strings.TrimSpace(req.RootPath)

	if req.Name == "" {
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "name is required",
		})

		return
	}

	if req.RootPath == "" {
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "root_path is required",
		})

		return
	}

	rootPath, err := filepath.Abs(req.RootPath)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "root_path is invalid",
		})

		return
	}

	info, err := os.Stat(rootPath)

	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			w.WriteHeader(http.StatusBadRequest)

			json.NewEncoder(w).Encode(map[string]any{
				"error": "root_path does not exist",
			})

			return
		}

		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "root_path is not accessible",
		})

		return
	}

	if !info.IsDir() {
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "root_path is not a directory",
		})

		return
	}

	id, err := store.CreateProject(r.Context(), req.Name, rootPath)

	if err != nil {
		if store.IsUniqueViolation(err) {
			w.WriteHeader(http.StatusConflict)

			json.NewEncoder(w).Encode(map[string]any{
				"error": "a project with this root_path already exists",
			})

			return
		}

		log.Printf("projects: create project: %v", err)

		w.WriteHeader(http.StatusInternalServerError)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "failed to create project",
		})

		return
	}

	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(models.Project{
		ID:       id,
		Name:     req.Name,
		RootPath: rootPath,
	})
}
