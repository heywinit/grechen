package cli

import (
	"fmt"
	"time"
)

// HandleToday shows situational awareness for today
func (c *CLI) HandleToday() error {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	// Get today's stats
	stats, err := c.stats.ComputeDailyStats(today)
	if err != nil {
		return err
	}

	fmt.Printf("today (%s)\n", today.Format("2006-01-02"))
	fmt.Printf("  logs: %d\n", stats.LogCount)
	if stats.WorkStartTime != nil {
		fmt.Printf("  work started: %s\n", stats.WorkStartTime.Format("15:04"))
	}
	fmt.Printf("  progress entries: %d\n", stats.ProgressEntries)
	fmt.Printf("  commitment updates: %d\n", stats.CommitmentUpdates)

	// Show open commitments
	commitments, err := c.store.ListOpenCommitments()
	if err != nil {
		return err
	}

	if len(commitments) > 0 {
		fmt.Println("\nopen commitments:")
		for _, c := range commitments {
			daysUntil := int(c.Expectation.Deadline.Sub(now).Hours() / 24)
			fmt.Printf("  %s → %s (due %s, %d days)\n",
				c.PersonID,
				c.Expectation.Description,
				c.Expectation.Deadline.Format("2006-01-02"),
				daysUntil)
		}
	}

	return nil
}

// HandleCommitments shows all commitments
func (c *CLI) HandleCommitments() error {
	commitments, err := c.store.ListCommitments()
	if err != nil {
		return err
	}

	if len(commitments) == 0 {
		fmt.Println("no commitments")
		return nil
	}

	fmt.Println("commitments:")
	for _, c := range commitments {
		fmt.Printf("  [%s] %s → %s (due %s, status: %s)\n",
			c.ID,
			c.PersonID,
			c.Expectation.Description,
			c.Expectation.Deadline.Format("2006-01-02"),
			c.Status)
	}

	return nil
}

// HandleReview shows stats summary and pattern alerts
func (c *CLI) HandleReview() error {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	// Get rolling stats (last 7 days)
	rollingStats, err := c.stats.ComputeRollingStats(today, 7)
	if err != nil {
		return err
	}

	fmt.Println("last 7 days:")
	fmt.Printf("  avg logs/day: %.1f\n", rollingStats.AvgLogCount)
	if rollingStats.AvgWorkStartTime != nil {
		fmt.Printf("  avg work start: %s\n", rollingStats.AvgWorkStartTime.Format("15:04"))
	}
	fmt.Printf("  avg progress entries/day: %.1f\n", rollingStats.AvgProgressEntries)
	fmt.Printf("  avg commitment updates/day: %.1f\n", rollingStats.AvgCommitmentUpdates)

	// Check patterns
	deviations, err := c.patterns.Evaluate(today, rollingStats)
	if err != nil {
		return err
	}

	if len(deviations) > 0 {
		fmt.Println("\npatterns:")
		for _, dev := range deviations {
			fmt.Printf("  [%s] %s: %s\n", dev.Severity, dev.Pattern, dev.Question.Text)
		}
	}

	return nil
}

// HandleThatsWrong handles correction flow
func (c *CLI) HandleThatsWrong() error {
	fmt.Println("what was wrong?")
	// In future: read from stdin or use Charm form
	// For now, this is a placeholder
	return fmt.Errorf("correction flow not fully implemented")
}
