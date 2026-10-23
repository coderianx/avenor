package models

// AddNewProjectReq is the request body for creating a project.
type AddNewProjectReq struct {
	Name     string `json:"name"`
	RootPath string `json:"root_path"`
}

// Project is the project representation returned by the API.
type Project struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	RootPath string `json:"root_path"`
}
