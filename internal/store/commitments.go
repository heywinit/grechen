package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/heywinit/grechen/internal/core"
)

const commitmentsFile = "commitments.json"

func (s *Store) SaveCommitment(commitment *core.Commitment) error {
	commitments, err := s.loadCommitments()
	if err != nil {
		return err
	}

	// Update or add commitment
	found := false
	for i, c := range commitments {
		if c.ID == commitment.ID {
			commitments[i] = commitment
			found = true
			break
		}
	}
	if !found {
		commitments = append(commitments, commitment)
	}

	return s.saveCommitments(commitments)
}

func (s *Store) GetCommitment(id string) (*core.Commitment, error) {
	commitments, err := s.loadCommitments()
	if err != nil {
		return nil, err
	}

	for _, c := range commitments {
		if c.ID == id {
			return c, nil
		}
	}

	return nil, fmt.Errorf("commitment not found: %s", id)
}

func (s *Store) ListCommitments() ([]*core.Commitment, error) {
	return s.loadCommitments()
}

func (s *Store) ListOpenCommitments() ([]*core.Commitment, error) {
	all, err := s.loadCommitments()
	if err != nil {
		return nil, err
	}

	var open []*core.Commitment
	for _, c := range all {
		if c.Status == core.StatusOpen || c.Status == core.StatusUpdated {
			open = append(open, c)
		}
	}

	return open, nil
}

func (s *Store) ListCommitmentsByPerson(personID string) ([]*core.Commitment, error) {
	all, err := s.loadCommitments()
	if err != nil {
		return nil, err
	}

	var filtered []*core.Commitment
	for _, c := range all {
		if c.PersonID == personID {
			filtered = append(filtered, c)
		}
	}

	return filtered, nil
}

func (s *Store) ListCommitmentsByProject(projectID string) ([]*core.Commitment, error) {
	all, err := s.loadCommitments()
	if err != nil {
		return nil, err
	}

	var filtered []*core.Commitment
	for _, c := range all {
		if c.ProjectID == projectID {
			filtered = append(filtered, c)
		}
	}

	return filtered, nil
}

func (s *Store) ListCommitmentsDueBefore(deadline time.Time) ([]*core.Commitment, error) {
	all, err := s.loadCommitments()
	if err != nil {
		return nil, err
	}

	var filtered []*core.Commitment
	for _, c := range all {
		if c.Expectation.Deadline.Before(deadline) && (c.Status == core.StatusOpen || c.Status == core.StatusUpdated) {
			filtered = append(filtered, c)
		}
	}

	return filtered, nil
}

func (s *Store) loadCommitments() ([]*core.Commitment, error) {
	filename := filepath.Join(s.MetaDir(), commitmentsFile)
	data, err := os.ReadFile(filename)
	if os.IsNotExist(err) {
		return []*core.Commitment{}, nil
	}
	if err != nil {
		return nil, err
	}

	var commitments []*core.Commitment
	if len(data) == 0 {
		return commitments, nil
	}

	if err := json.Unmarshal(data, &commitments); err != nil {
		return nil, fmt.Errorf("failed to unmarshal commitments: %w", err)
	}

	return commitments, nil
}

func (s *Store) saveCommitments(commitments []*core.Commitment) error {
	filename := filepath.Join(s.MetaDir(), commitmentsFile)
	data, err := json.MarshalIndent(commitments, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal commitments: %w", err)
	}

	return os.WriteFile(filename, data, 0644)
}
