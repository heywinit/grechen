package rules

import (
	"fmt"
	"time"

	"github.com/heywinit/grechen/internal/core"
)

func (r *Rules) validateUpdate(candidate core.Candidate, entry *core.Entry) (*ValidationResult, error) {
	var questions []core.Question

	// Try to find commitment by ID first
	commitmentID, _ := candidate.Data["commitment_id"].(string)
	var commitment *core.Commitment
	var err error

	if commitmentID != "" {
		commitment, err = r.store.GetCommitment(commitmentID)
		if err != nil {
			questions = append(questions, core.Question{
				ID:       "commitment_id",
				Text:     fmt.Sprintf("commitment %s not found. which commitment?", commitmentID),
				Required: true,
				Field:    "commitment_id",
			})
		}
	} else {
		// Try to find by person/project
		personID, _ := candidate.Data["person"].(string)
		projectID, _ := candidate.Data["project"].(string)

		if personID == "" && projectID == "" {
			questions = append(questions, core.Question{
				ID:       "commitment_identifier",
				Text:     "which commitment is this updating?",
				Required: true,
				Field:    "commitment_id",
			})
		} else {
			// Find matching commitment
			var commitments []*core.Commitment
			if personID != "" {
				commitments, err = r.store.ListCommitmentsByPerson(personID)
			} else if projectID != "" {
				commitments, err = r.store.ListCommitmentsByProject(projectID)
			}

			if err != nil {
				return nil, fmt.Errorf("failed to find commitments: %w", err)
			}

			// Filter to open/updated commitments
			var open []*core.Commitment
			for _, c := range commitments {
				if c.Status == core.StatusOpen || c.Status == core.StatusUpdated {
					open = append(open, c)
				}
			}

			if len(open) == 0 {
				questions = append(questions, core.Question{
					ID:       "commitment_not_found",
					Text:     "no open commitments found. which commitment?",
					Required: true,
					Field:    "commitment_id",
				})
			} else if len(open) == 1 {
				commitment = open[0]
			} else {
				questions = append(questions, core.Question{
					ID:       "commitment_ambiguous",
					Text:     fmt.Sprintf("multiple commitments found (%d). which one?", len(open)),
					Required: true,
					Field:    "commitment_id",
				})
			}
		}
	}

	if len(questions) > 0 {
		return &ValidationResult{
			Valid:     false,
			Questions: questions,
		}, nil
	}

	// Determine status update
	status := core.StatusUpdated
	description, _ := candidate.Data["status"].(string)
	if description == "done" || description == "completed" || description == "finished" {
		status = core.StatusFulfilled
	}

	update := &CommitmentUpdate{
		CommitmentID: commitment.ID,
		Status:       status,
		Description:  description,
	}

	return &ValidationResult{
		Valid: true,
		Action: Action{
			Type:   core.IntentUpdate,
			Entry:  entry,
			Update: update,
		},
	}, nil
}

func (r *Rules) validateProgress(candidate core.Candidate, entry *core.Entry) (*ValidationResult, error) {
	projectID, ok := candidate.Data["project"].(string)
	if !ok || projectID == "" {
		return &ValidationResult{
			Valid: false,
			Questions: []core.Question{
				{
					ID:       "project",
					Text:     "which project is this progress for?",
					Required: true,
					Field:    "project",
				},
			},
		}, nil
	}

	// Check if project exists, create if not
	_, err := r.store.GetProject(projectID)
	if err != nil {
		project := &core.Project{
			ID:       projectID,
			Priority: 0,
			Metadata: make(map[string]any),
		}
		if err := r.store.SaveProject(project); err != nil {
			return nil, fmt.Errorf("failed to save project: %w", err)
		}
	}

	status, _ := candidate.Data["status"].(string)
	notes, _ := candidate.Data["notes"].(string)

	return &ValidationResult{
		Valid: true,
		Action: Action{
			Type: core.IntentProgress,
			Entry: entry,
			Progress: &Progress{
				ProjectID: projectID,
				Status:    status,
				Notes:     notes,
			},
		},
	}, nil
}

func (r *Rules) validateEvent(candidate core.Candidate, entry *core.Entry) (*ValidationResult, error) {
	var questions []core.Question

	timeStr, ok := candidate.Data["time"].(string)
	if !ok || timeStr == "" {
		questions = append(questions, core.Question{
			ID:       "time",
			Text:     "when is this event?",
			Required: true,
			Field:    "time",
		})
	}

	if len(questions) > 0 {
		return &ValidationResult{
			Valid:     false,
			Questions: questions,
		}, nil
	}

	// Parse time
	eventTime, err := time.Parse("2006-01-02 15:04", timeStr)
	if err != nil {
		// Try other formats
		eventTime, err = time.Parse("2006-01-02", timeStr)
		if err != nil {
			return nil, fmt.Errorf("invalid time format: %w", err)
		}
	}

	personID, _ := candidate.Data["person"].(string)
	projectID, _ := candidate.Data["project"].(string)
	title, _ := candidate.Data["title"].(string)

	return &ValidationResult{
		Valid: true,
		Action: Action{
			Type: core.IntentEvent,
			Entry: entry,
			Event: &Event{
				Time:      eventTime,
				PersonID:  personID,
				ProjectID: projectID,
				Title:     title,
			},
		},
	}, nil
}
