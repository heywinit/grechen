package cli

import (
	"github.com/heywinit/grechen/internal/core"
)

// HandleInteractive handles interactive question answering
// Future: implement with Charm forms for better UX
func (c *CLI) HandleInteractive(questions []core.Question) (map[string]string, error) {
	// Placeholder for future Charm-based interactive forms
	// For now, returns empty map
	answers := make(map[string]string)
	return answers, nil
}
