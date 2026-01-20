package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/heywinit/grechen/internal/core"
)

const peopleFile = "people.json"

func (s *Store) SavePerson(person *core.Person) error {
	people, err := s.loadPeople()
	if err != nil {
		return err
	}

	// Update or add person
	found := false
	for i, p := range people {
		if p.ID == person.ID {
			people[i] = person
			found = true
			break
		}
	}
	if !found {
		people = append(people, person)
	}

	return s.savePeople(people)
}

func (s *Store) GetPerson(id string) (*core.Person, error) {
	people, err := s.loadPeople()
	if err != nil {
		return nil, err
	}

	for _, p := range people {
		if p.ID == id {
			return p, nil
		}
	}

	return nil, fmt.Errorf("person not found: %s", id)
}

func (s *Store) ListPeople() ([]*core.Person, error) {
	return s.loadPeople()
}

func (s *Store) FindPersonByName(name string) (*core.Person, error) {
	people, err := s.loadPeople()
	if err != nil {
		return nil, err
	}

	// Case-insensitive search
	nameLower := strings.ToLower(name)
	for _, p := range people {
		if strings.ToLower(p.Name) == nameLower || strings.ToLower(p.ID) == nameLower {
			return p, nil
		}
	}

	return nil, fmt.Errorf("person not found: %s", name)
}

func (s *Store) loadPeople() ([]*core.Person, error) {
	filename := filepath.Join(s.MetaDir(), peopleFile)
	data, err := os.ReadFile(filename)
	if os.IsNotExist(err) {
		return []*core.Person{}, nil
	}
	if err != nil {
		return nil, err
	}

	var people []*core.Person
	if len(data) == 0 {
		return people, nil
	}

	if err := json.Unmarshal(data, &people); err != nil {
		return nil, fmt.Errorf("failed to unmarshal people: %w", err)
	}

	return people, nil
}

func (s *Store) savePeople(people []*core.Person) error {
	filename := filepath.Join(s.MetaDir(), peopleFile)
	data, err := json.MarshalIndent(people, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal people: %w", err)
	}

	return os.WriteFile(filename, data, 0644)
}
