package cli

import (
	"fmt"
	"time"

	"github.com/briandowns/spinner"
	"github.com/heywinit/grechen/internal/core"
	"github.com/heywinit/grechen/internal/extract"
	"github.com/heywinit/grechen/internal/patterns"
	"github.com/heywinit/grechen/internal/rules"
	"github.com/heywinit/grechen/internal/stats"
	"github.com/heywinit/grechen/internal/store"
)

type CLI struct {
	store    *store.Store
	extractor extract.Extractor
	rules    *rules.Rules
	stats    *stats.Stats
	patterns *patterns.Patterns
}

func New(s *store.Store, ext extract.Extractor, r *rules.Rules, st *stats.Stats, p *patterns.Patterns) *CLI {
	return &CLI{
		store:    s,
		extractor: ext,
		rules:    r,
		stats:    st,
		patterns: p,
	}
}

// HandleInput processes natural language input
func (c *CLI) HandleInput(input string) error {
	// Create entry
	entry := &core.Entry{
		ID:        generateID(),
		Timestamp: time.Now(),
		Raw:       input,
	}

	// Show spinner while LLM is processing
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Prefix = "processing "
	s.Start()

	// Extract candidates
	candidates, questions, err := c.extractor.Extract(input)
	
	s.Stop()
	
	if err != nil {
		return fmt.Errorf("extraction failed: %w", err)
	}

	// If we have blocking questions from extraction, ask them
	if len(questions) > 0 {
		return c.handleQuestions(questions, entry)
	}

	// Process each candidate
	for _, candidate := range candidates {
		if err := c.processCandidate(candidate, entry); err != nil {
			return err
		}
	}

	return nil
}

func (c *CLI) processCandidate(candidate core.Candidate, entry *core.Entry) error {
	// Validate with rules
	result, err := c.rules.Validate(candidate, entry)
	if err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// If we have blocking questions, ask them
	if !result.Valid {
		return c.handleQuestions(result.Questions, entry)
	}

	// Execute action
	return c.executeAction(result.Action, candidate, entry)
}

func (c *CLI) executeAction(action rules.Action, candidate core.Candidate, entry *core.Entry) error {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	switch action.Type {
	case core.IntentLog:
		if err := c.store.AppendLog(today, entry); err != nil {
			return fmt.Errorf("failed to append log: %w", err)
		}
		fmt.Printf("logged as: %s (confidence: %.2f)\n", candidate.Type, candidate.Confidence)

	case core.IntentCommitment:
		if err := c.store.SaveCommitment(action.Commitment); err != nil {
			return fmt.Errorf("failed to save commitment: %w", err)
		}
		if err := c.store.AppendCommitment(today, action.Commitment); err != nil {
			return fmt.Errorf("failed to append commitment: %w", err)
		}
		fmt.Printf("logged commitment to %s: %s (due %s, confidence: %.2f)\n",
			action.Commitment.PersonID,
			action.Commitment.Expectation.Description,
			action.Commitment.Expectation.Deadline.Format("2006-01-02"),
			candidate.Confidence)

	case core.IntentUpdate:
		commitment, err := c.store.GetCommitment(action.Update.CommitmentID)
		if err != nil {
			return fmt.Errorf("commitment not found: %w", err)
		}

		// Update commitment
		commitment.Status = action.Update.Status
		now := time.Now()
		commitment.LastUpdateAt = &now
		commitment.History = append(commitment.History, core.CommitmentEvent{
			Timestamp:   now,
			Type:        string(action.Update.Status),
			Description: action.Update.Description,
		})

		if err := c.store.SaveCommitment(commitment); err != nil {
			return fmt.Errorf("failed to update commitment: %w", err)
		}
		if err := c.store.AppendCommitment(today, commitment); err != nil {
			return fmt.Errorf("failed to append commitment update: %w", err)
		}
		fmt.Printf("updated commitment: %s\n", commitment.Expectation.Description)

	case core.IntentProgress:
		if err := c.store.AppendLog(today, entry); err != nil {
			return fmt.Errorf("failed to append progress: %w", err)
		}
		fmt.Printf("logged progress on %s (confidence: %.2f)\n", action.Progress.ProjectID, candidate.Confidence)

	case core.IntentEvent:
		// For now, just log the event
		if err := c.store.AppendLog(today, entry); err != nil {
			return fmt.Errorf("failed to append event: %w", err)
		}
		fmt.Printf("logged event: %s (confidence: %.2f)\n", action.Event.Title, candidate.Confidence)

	case core.IntentCorrection:
		if err := c.store.AppendNote(today, fmt.Sprintf("correction: %s", entry.Raw)); err != nil {
			return fmt.Errorf("failed to append correction: %w", err)
		}
		fmt.Printf("correction logged (confidence: %.2f)\n", candidate.Confidence)

	default:
		return fmt.Errorf("unknown action type: %s", action.Type)
	}

	return nil
}

func (c *CLI) handleQuestions(questions []core.Question, entry *core.Entry) error {
	// For now, just print questions (future: interactive Charm forms)
	for _, q := range questions {
		fmt.Printf("? %s\n", q.Text)
	}
	return fmt.Errorf("questions need answers")
}

func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
