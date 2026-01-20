package cli

import (
	"fmt"
	"time"

	"github.com/heywinit/grechen/internal/patterns"
)

// HandleGoodnight implements the goodnight routine
func (c *CLI) HandleGoodnight() error {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	// Get today's stats
	todayStats, err := c.stats.ComputeDailyStats(today)
	if err != nil {
		return err
	}

	// Get rolling stats for comparison (last 7 days)
	rollingStats, err := c.stats.ComputeRollingStats(today, 7)
	if err != nil {
		return err
	}

	// Evaluate patterns
	deviations, err := c.patterns.Evaluate(today, rollingStats)
	if err != nil {
		return err
	}

	// Generate questions (1-5)
	questions := patterns.GenerateQuestions(deviations, 5)

	// Compare today to ideal
	fmt.Println("goodnight")
	fmt.Printf("today: %d logs, ", todayStats.LogCount)
	if todayStats.WorkStartTime != nil {
		fmt.Printf("started at %s, ", todayStats.WorkStartTime.Format("15:04"))
	}
	fmt.Printf("%d progress entries\n", todayStats.ProgressEntries)

	if rollingStats.Days > 0 {
		fmt.Printf("vs avg: %.1f logs/day", rollingStats.AvgLogCount)
		if rollingStats.AvgWorkStartTime != nil {
			fmt.Printf(", start at %s", rollingStats.AvgWorkStartTime.Format("15:04"))
		}
		fmt.Println()
	}

	// Ask questions
	if len(questions) > 0 {
		fmt.Println("\nquestions:")
		for i, q := range questions {
			fmt.Printf("%d. %s\n", i+1, q.Text)
		}

		// For now, just append questions to notes
		// In future: interactive response collection
		notes := fmt.Sprintf("goodnight questions:\n")
		for i, q := range questions {
			notes += fmt.Sprintf("%d. %s\n", i+1, q.Text)
		}
		if err := c.store.AppendNote(today, notes); err != nil {
			return fmt.Errorf("failed to append questions: %w", err)
		}
	} else {
		fmt.Println("\nno questions today")
	}

	return nil
}
