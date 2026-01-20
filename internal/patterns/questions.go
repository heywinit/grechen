package patterns

import (
	"github.com/heywinit/grechen/internal/core"
)

// GenerateQuestions converts deviations into questions, limiting to 1-5 questions
func GenerateQuestions(deviations []core.Deviation, maxQuestions int) []core.Question {
	if maxQuestions <= 0 {
		maxQuestions = 5
	}

	// Simple selection: take highest severity first
	selected := make([]core.Question, 0, maxQuestions)
	added := make(map[string]bool)

	// First pass: high severity
	for _, dev := range deviations {
		if dev.Severity == "high" && !added[dev.Question.ID] && len(selected) < maxQuestions {
			selected = append(selected, dev.Question)
			added[dev.Question.ID] = true
		}
	}

	// Second pass: medium severity
	for _, dev := range deviations {
		if dev.Severity == "medium" && !added[dev.Question.ID] && len(selected) < maxQuestions {
			selected = append(selected, dev.Question)
			added[dev.Question.ID] = true
		}
	}

	// Third pass: low severity
	for _, dev := range deviations {
		if dev.Severity == "low" && !added[dev.Question.ID] && len(selected) < maxQuestions {
			selected = append(selected, dev.Question)
			added[dev.Question.ID] = true
		}
	}

	return selected
}
