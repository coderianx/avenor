package store

import (
	"context"
	"testing"
)

func TestCreateAndListProjects(t *testing.T) {
	openTestDB(t)

	ctx := context.Background()

	empty, err := ListProjects(ctx)

	if err != nil {
		t.Fatalf("ListProjects on empty database: %v", err)
	}

	if empty == nil {
		t.Fatal("ListProjects must return an empty slice, not nil")
	}

	if len(empty) != 0 {
		t.Fatalf("projects = %d, want 0", len(empty))
	}

	rootPath := t.TempDir()

	id, err := CreateProject(ctx, "first", rootPath)

	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}

	projects, err := ListProjects(ctx)

	if err != nil {
		t.Fatalf("ListProjects: %v", err)
	}

	if len(projects) != 1 {
		t.Fatalf("projects = %d, want 1", len(projects))
	}

	got := projects[0]

	if got.ID != id || got.Name != "first" || got.RootPath != rootPath {
		t.Fatalf("project = %+v, want id=%d name=first root_path=%s", got, id, rootPath)
	}

	// A duplicate root_path must be rejected.
	_, err = CreateProject(ctx, "second", rootPath)

	if !IsUniqueViolation(err) {
		t.Fatalf("duplicate root_path error = %v, want unique violation", err)
	}
}

func TestDeleteProject(t *testing.T) {
	openTestDB(t)

	ctx := context.Background()

	id, err := CreateProject(ctx, "first", t.TempDir())

	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}

	deleted, err := DeleteProject(ctx, id)

	if err != nil {
		t.Fatalf("DeleteProject: %v", err)
	}

	if !deleted {
		t.Fatal("DeleteProject = false, want true")
	}

	projects, err := ListProjects(ctx)

	if err != nil {
		t.Fatalf("ListProjects: %v", err)
	}

	if len(projects) != 0 {
		t.Fatalf("projects = %d, want 0", len(projects))
	}

	// Deleting a missing project must not report a deletion.
	deleted, err = DeleteProject(ctx, id)

	if err != nil {
		t.Fatalf("DeleteProject missing: %v", err)
	}

	if deleted {
		t.Fatal("DeleteProject on missing project = true, want false")
	}
}
