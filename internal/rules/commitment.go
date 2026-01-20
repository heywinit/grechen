package rules

import (
	"fmt"
	"time"

	"github.com/heywinit/grechen/internal/core"
)

func (r *Rules) validateCommitment(candidate core.Candidate, entry *core.Entry) (*ValidationResult, error) {
	var questions []core.Question

	// Extract person
	personID, ok := candidate.Data["person"].(string)
	if !ok {
		questions = append(questions, core.Question{
			ID:       "person",
			Text:     "who is this commitment to?",
			Required: true,
			Field:    "person",
		})
	} else {
		// Check if person exists, if not create placeholder
		_, err := r.store.GetPerson(personID)
		if err != nil {
			// Person doesn't exist, create it
			person := &core.Person{
				ID:       personID,
				Name:     personID,
				Metadata: make(map[string]any),
			}
			if err := r.store.SavePerson(person); err != nil {
				return nil, fmt.Errorf("failed to save person: %w", err)
			}
		}
	}

	// Extract expectation
	expRaw, ok := candidate.Data["expectation"].(map[string]any)
	if !ok {
		questions = append(questions, core.Question{
			ID:       "expectation",
			Text:     "what is the commitment?",
			Required: true,
			Field:    "expectation",
		})
	} else {
		description, _ := expRaw["description"].(string)
		if description == "" {
			questions = append(questions, core.Question{
				ID:       "expectation_description",
				Text:     "what exactly is the commitment?",
				Required: true,
				Field:    "expectation.description",
			})
		}

		deadlineStr, _ := expRaw["deadline"].(string)
		if deadlineStr == "" {
			questions = append(questions, core.Question{
				ID:       "expectation_deadline",
				Text:     "when is this due?",
				Required: true,
				Field:    "expectation.deadline",
			})
		}
	}

	// Extract project (optional)
	projectID, _ := candidate.Data["project"].(string)
	if projectID != "" {
		_, err := r.store.GetProject(projectID)
		if err != nil {
			// Project doesn't exist, create it
			project := &core.Project{
				ID:       projectID,
				Priority: 0,
				Metadata: make(map[string]any),
			}
			if err := r.store.SaveProject(project); err != nil {
				return nil, fmt.Errorf("failed to save project: %w", err)
			}
		}
	}

	// If we have blocking questions, return them
	if len(questions) > 0 {
		return &ValidationResult{
			Valid:     false,
			Questions: questions,
		}, nil
	}

	// Build commitment
	expRaw = candidate.Data["expectation"].(map[string]any)
	description := expRaw["description"].(string)
	deadlineStr := expRaw["deadline"].(string)
	deadline, err := time.Parse("2006-01-02", deadlineStr)
	if err != nil {
		return nil, fmt.Errorf("invalid deadline: %w", err)
	}

	hardness := "soft"
	if h, ok := expRaw["hardness"].(string); ok {
		hardness = h
	}

	commitment := &core.Commitment{
		ID:          generateID(),
		CreatedAt:   time.Now(),
		SourceEntry: entry.ID,
		PersonID:    personID,
		ProjectID:   projectID,
		Expectation: core.Expectation{
			Description: description,
			Deadline:    deadline,
			Hardness:    hardness,
		},
		Status:  core.StatusOpen,
		History: []core.CommitmentEvent{},
	}

	return &ValidationResult{
		Valid: true,
		Action: Action{
			Type:       core.IntentCommitment,
			Entry:      entry,
			Commitment: commitment,
		},
	}, nil
}

func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
