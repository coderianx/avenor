package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/coderianx/avenor/internal/models"
)

// CreateProject stores a project and returns its id.
func CreateProject(
	ctx context.Context,
	name string,
	rootPath string,
) (int64, error) {
	if DB == nil {
		return 0, errors.New("database not connected")
	}

	result, err := DB.ExecContext(
		ctx,
		`INSERT INTO projects (name, root_path) VALUES (?, ?)`,
		name,
		rootPath,
	)

	if err != nil {
		return 0, fmt.Errorf("insert project: %w", err)
	}

	id, err := result.LastInsertId()

	if err != nil {
		return 0, fmt.Errorf("project id: %w", err)
	}

	return id, nil
}

// DeleteProject removes the project with the given id and reports whether a
// project was actually deleted.
func DeleteProject(
	ctx context.Context,
	id int64,
) (bool, error) {
	if DB == nil {
		return false, errors.New("database not connected")
	}

	result, err := DB.ExecContext(
		ctx,
		`DELETE FROM projects WHERE id = ?`,
		id,
	)

	if err != nil {
		return false, fmt.Errorf("delete project: %w", err)
	}

	affected, err := result.RowsAffected()

	if err != nil {
		return false, fmt.Errorf("deleted project rows: %w", err)
	}

	return affected > 0, nil
}

// ListProjects returns every project ordered by id.
func ListProjects(ctx context.Context) ([]models.Project, error) {
	if DB == nil {
		return nil, errors.New("database not connected")
	}

	rows, err := DB.QueryContext(
		ctx,
		`SELECT id, name, root_path FROM projects ORDER BY id`,
	)

	if err != nil {
		return nil, fmt.Errorf("list projects: %w", err)
	}

	defer rows.Close()

	projects := make([]models.Project, 0)

	for rows.Next() {
		var project models.Project

		err := rows.Scan(&project.ID, &project.Name, &project.RootPath)

		if err != nil {
			return nil, fmt.Errorf("scan project: %w", err)
		}

		projects = append(projects, project)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate projects: %w", err)
	}

	return projects, nil
}
