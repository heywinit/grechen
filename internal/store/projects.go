package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/heywinit/grechen/internal/core"
)

const projectsFile = "projects.json"

func (s *Store) SaveProject(project *core.Project) error {
	projects, err := s.loadProjects()
	if err != nil {
		return err
	}

	// Update or add project
	found := false
	for i, p := range projects {
		if p.ID == project.ID {
			projects[i] = project
			found = true
			break
		}
	}
	if !found {
		projects = append(projects, project)
	}

	return s.saveProjects(projects)
}

func (s *Store) GetProject(id string) (*core.Project, error) {
	projects, err := s.loadProjects()
	if err != nil {
		return nil, err
	}

	for _, p := range projects {
		if p.ID == id {
			return p, nil
		}
	}

	return nil, fmt.Errorf("project not found: %s", id)
}

func (s *Store) ListProjects() ([]*core.Project, error) {
	return s.loadProjects()
}

func (s *Store) loadProjects() ([]*core.Project, error) {
	filename := filepath.Join(s.MetaDir(), projectsFile)
	data, err := os.ReadFile(filename)
	if os.IsNotExist(err) {
		return []*core.Project{}, nil
	}
	if err != nil {
		return nil, err
	}

	var projects []*core.Project
	if len(data) == 0 {
		return projects, nil
	}

	if err := json.Unmarshal(data, &projects); err != nil {
		return nil, fmt.Errorf("failed to unmarshal projects: %w", err)
	}

	return projects, nil
}

func (s *Store) saveProjects(projects []*core.Project) error {
	filename := filepath.Join(s.MetaDir(), projectsFile)
	data, err := json.MarshalIndent(projects, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal projects: %w", err)
	}

	return os.WriteFile(filename, data, 0644)
}
