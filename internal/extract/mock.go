package extract

import (
	"strings"
	"time"

	"github.com/heywinit/grechen/internal/core"
)

// MockExtractor is a simple mock implementation for testing
type MockExtractor struct {
	responses map[string]MockResponse
}

type MockResponse struct {
	Candidates []core.Candidate
	Questions  []core.Question
	Error      error
}

func NewMockExtractor() *MockExtractor {
	return &MockExtractor{
		responses: make(map[string]MockResponse),
	}
}

// SetResponse allows setting a mock response for a given input
func (m *MockExtractor) SetResponse(input string, candidates []core.Candidate, questions []core.Question, err error) {
	m.responses[strings.ToLower(input)] = MockResponse{
		Candidates: candidates,
		Questions:  questions,
		Error:      err,
	}
}

func (m *MockExtractor) Extract(input string) ([]core.Candidate, []core.Question, error) {
	// Check for predefined response
	if resp, ok := m.responses[strings.ToLower(input)]; ok {
		return resp.Candidates, resp.Questions, resp.Error
	}

	// Default behavior: simple pattern matching
	inputLower := strings.ToLower(input)

	// Check for commitment patterns
	if strings.Contains(inputLower, "told") || strings.Contains(inputLower, "promised") || strings.Contains(inputLower, "will") {
		// Try to extract commitment
		candidate := core.Candidate{
			Type:       core.IntentCommitment,
			Confidence: 0.85,
			Data:       make(map[string]any),
		}

		// Simple extraction (mock)
		if strings.Contains(inputLower, "prof") {
			candidate.Data["person"] = "prof"
		}
		if strings.Contains(inputLower, "tomorrow") {
			tomorrow := time.Now().AddDate(0, 0, 1)
			candidate.Data["expectation"] = map[string]any{
				"description": extractDescription(input),
				"deadline":    tomorrow.Format("2006-01-02"),
				"hardness":    "hard",
			}
		}

		if len(candidate.Data) > 0 {
			return []core.Candidate{candidate}, []core.Question{}, nil
		}
	}

	// Check for progress patterns
	if strings.Contains(inputLower, "done") || strings.Contains(inputLower, "finished") || strings.Contains(inputLower, "completed") {
		candidate := core.Candidate{
			Type:       core.IntentProgress,
			Confidence: 0.80,
			Data: map[string]any{
				"project": extractProject(input),
			},
		}
		return []core.Candidate{candidate}, []core.Question{}, nil
	}

	// Default: log entry
	candidate := core.Candidate{
		Type:       core.IntentLog,
		Confidence: 0.90,
		Data: map[string]any{
			"text": input,
		},
	}
	return []core.Candidate{candidate}, []core.Question{}, nil
}

func extractDescription(input string) string {
	// Simple extraction - in real implementation, LLM would do this
	words := strings.Fields(input)
	if len(words) > 3 {
		return strings.Join(words[len(words)-3:], " ")
	}
	return input
}

func extractProject(input string) string {
	inputLower := strings.ToLower(input)
	projects := []string{"caresad", "farmer", "kaifu", "soldecoder"}
	for _, p := range projects {
		if strings.Contains(inputLower, p) {
			return p
		}
	}
	return "unknown"
}
