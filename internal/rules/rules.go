package rules

import (
	"fmt"
	"time"

	"github.com/heywinit/grechen/internal/core"
	"github.com/heywinit/grechen/internal/store"
)

type Rules struct {
	store *store.Store
}

func New(s *store.Store) *Rules {
	return &Rules{store: s}
}

// ValidateCandidate validates a candidate and returns either a validated action or blocking questions
type ValidationResult struct {
	Valid   bool
	Action  Action
	Questions []core.Question
}

type Action struct {
	Type       core.IntentType
	Entry      *core.Entry
	Commitment *core.Commitment
	Update     *CommitmentUpdate
	Event      *Event
	Progress   *Progress
}

type CommitmentUpdate struct {
	CommitmentID string
	Status       core.CommitmentStatus
	Description  string
}

type Event struct {
	Time      time.Time
	PersonID  string
	ProjectID string
	Title     string
}

type Progress struct {
	ProjectID string
	Status    string
	Notes     string
}

func (r *Rules) Validate(candidate core.Candidate, entry *core.Entry) (*ValidationResult, error) {
	switch candidate.Type {
	case core.IntentCommitment:
		return r.validateCommitment(candidate, entry)
	case core.IntentUpdate:
		return r.validateUpdate(candidate, entry)
	case core.IntentProgress:
		return r.validateProgress(candidate, entry)
	case core.IntentEvent:
		return r.validateEvent(candidate, entry)
	case core.IntentLog, core.IntentCorrection:
		return &ValidationResult{
			Valid: true,
			Action: Action{
				Type:  candidate.Type,
				Entry: entry,
			},
		}, nil
	default:
		return nil, fmt.Errorf("unknown intent type: %s", candidate.Type)
	}
}
