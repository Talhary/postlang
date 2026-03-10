package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Node represents a folder or an endpoint in the tree
type Node struct {
	UID       string    `json:"uid"`
	ParentUID string    `json:"parent_uid"`
	Name      string    `json:"name"`
	IsFolder  bool      `json:"is_folder"`
	Endpoint  *Endpoint `json:"endpoint,omitempty"`
}

// Project represents a single workspace environment (APIs and Variables)
type Project struct {
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	Nodes     map[string]*Node `json:"nodes"`
	RootNodes []string        `json:"root_nodes"`
	Variables string          `json:"variables"`
}

// Workspace holds all projects
type Workspace struct {
	Projects []Project `json:"projects"`
}

func getWorkspacePath() string {
	home, err := os.UserConfigDir()
	if err != nil {
		home = "."
	}
	// Store state in the application's config directory
	appDir := filepath.Join(home, "Postlang")
	os.MkdirAll(appDir, 0755)
	
	return filepath.Join(appDir, "workspace.json")
}

// LoadWorkspace loads the workspace from the config file, generating a default if needed
func LoadWorkspace() (*Workspace, error) {
	path := getWorkspacePath()
	
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Return default workspace
			return &Workspace{
				Projects: []Project{
					{
						ID:        "default",
						Name:      "Default Project",
						Nodes:     make(map[string]*Node),
						RootNodes: []string{},
						Variables: "BASE_URL=https://httpbin.org\nTOKEN=my-secret-token",
					},
				},
			}, nil
		}
		return nil, err
	}

	var ws Workspace
	if err := json.Unmarshal(data, &ws); err != nil {
		return nil, err
	}
	
	if len(ws.Projects) == 0 {
		ws.Projects = append(ws.Projects, Project{
			ID:        "default",
			Name:      "Default Project",
			Nodes:     make(map[string]*Node),
			RootNodes: []string{},
			Variables: "",
		})
	}

	// Migration and safety for older versions
	for i := range ws.Projects {
		if ws.Projects[i].Nodes == nil {
			ws.Projects[i].Nodes = make(map[string]*Node)
		}
	}

	return &ws, nil
}

// SaveWorkspace persists the workspace to disk
func SaveWorkspace(ws *Workspace) error {
	path := getWorkspacePath()
	
	data, err := json.MarshalIndent(ws, "", "  ")
	if err != nil {
		return err
	}
	
	return os.WriteFile(path, data, 0644)
}
