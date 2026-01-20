package extract

import (
	"github.com/heywinit/grechen/internal/core"
)

// Extractor parses natural language input into structured candidates
type Extractor interface {
	Extract(input string) ([]core.Candidate, []core.Question, error)
}

const (
	MinConfidence = 0.7
)
