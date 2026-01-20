package patterns

import (
	"fmt"
	"time"

	"github.com/heywinit/grechen/internal/core"
	"github.com/heywinit/grechen/internal/stats"
)

const (
	idealWorkStartHour = 9 // 9 AM
	lateStartThreshold = 2 // 2 hours late
	sparseLogThreshold = 0.5 // 50% of average
)

func (p *Patterns) detectLateStart(todayStats *core.DailyStats, rollingStats *stats.RollingStats) *core.Deviation {
	if todayStats.WorkStartTime == nil {
		return nil // No work started today
	}

	workHour := todayStats.WorkStartTime.Hour()
	idealHour := idealWorkStartHour

	// Use rolling average if available
	if rollingStats.AvgWorkStartTime != nil {
		idealHour = rollingStats.AvgWorkStartTime.Hour()
	}

	if workHour > idealHour+lateStartThreshold {
		return &core.Deviation{
			Pattern:  core.PatternLateStart,
			Severity: "medium",
			Question: core.Question{
				ID:       "late_start",
				Text:     fmt.Sprintf("started work at %d:00, %d hours later than usual. what happened?", workHour, workHour-idealHour),
				Required: false,
				Field:    "work_start",
			},
		}
	}

	return nil
}

func (p *Patterns) detectSparseLogs(todayStats *core.DailyStats, rollingStats *stats.RollingStats) *core.Deviation {
	if rollingStats.Days == 0 {
		return nil // No baseline
	}

	avgLogs := rollingStats.AvgLogCount
	if avgLogs == 0 {
		return nil
	}

	ratio := float64(todayStats.LogCount) / avgLogs
	if ratio < sparseLogThreshold && todayStats.LogCount < int(avgLogs) {
		return &core.Deviation{
			Pattern:  core.PatternSparseLogs,
			Severity: "low",
			Question: core.Question{
				ID:       "sparse_logs",
				Text:     fmt.Sprintf("only %d log entries today (avg: %.1f). anything notable?", todayStats.LogCount, avgLogs),
				Required: false,
				Field:    "logs",
			},
		}
	}

	return nil
}

func (p *Patterns) detectCommitmentSilence(date time.Time) ([]core.Deviation, error) {
	commitments, err := p.store.ListOpenCommitments()
	if err != nil {
		return nil, err
	}

	var deviations []core.Deviation
	silenceThreshold := 3 * 24 * time.Hour // 3 days

	for _, commitment := range commitments {
		var lastUpdate time.Time
		if commitment.LastUpdateAt != nil {
			lastUpdate = *commitment.LastUpdateAt
		} else {
			lastUpdate = commitment.CreatedAt
		}

		daysSinceUpdate := date.Sub(lastUpdate)
		if daysSinceUpdate > silenceThreshold {
			days := int(daysSinceUpdate.Hours() / 24)
			deviations = append(deviations, core.Deviation{
				Pattern:  core.PatternCommitmentSilence,
				Severity: "high",
				Question: core.Question{
					ID:       fmt.Sprintf("commitment_silence_%s", commitment.ID),
					Text:     fmt.Sprintf("no update on commitment to %s (%s) in %d days. still on track?", commitment.PersonID, commitment.Expectation.Description, days),
					Required: false,
					Field:    "commitment_update",
				},
			})
		}
	}

	return deviations, nil
}

func (p *Patterns) detectRepeatedViolations() ([]core.Deviation, error) {
	commitments, err := p.store.ListCommitments()
	if err != nil {
		return nil, err
	}

	// Count violations per person
	violationCount := make(map[string]int)
	for _, c := range commitments {
		if c.Status == core.StatusViolated {
			violationCount[c.PersonID]++
		}
	}

	var deviations []core.Deviation
	for personID, count := range violationCount {
		if count >= 2 {
			deviations = append(deviations, core.Deviation{
				Pattern:  core.PatternRepeatedViolations,
				Severity: "high",
				Question: core.Question{
					ID:       fmt.Sprintf("repeated_violations_%s", personID),
					Text:     fmt.Sprintf("%d violated commitments with %s. pattern?", count, personID),
					Required: false,
					Field:    "violations",
				},
			})
		}
	}

	return deviations, nil
}

func (p *Patterns) detectOptimisticStall() ([]core.Deviation, error) {
	commitments, err := p.store.ListOpenCommitments()
	if err != nil {
		return nil, err
	}

	var deviations []core.Deviation
	now := time.Now()

	for _, commitment := range commitments {
		// Check if commitment has been updated multiple times but not fulfilled
		if commitment.Status == core.StatusUpdated && len(commitment.History) >= 2 {
			// Check if deadline is approaching or passed
			daysUntilDeadline := commitment.Expectation.Deadline.Sub(now).Hours() / 24
			if daysUntilDeadline < 1 && commitment.Status != core.StatusFulfilled {
				deviations = append(deviations, core.Deviation{
					Pattern:  core.PatternOptimisticStall,
					Severity: "medium",
					Question: core.Question{
						ID:       fmt.Sprintf("optimistic_stall_%s", commitment.ID),
						Text:     fmt.Sprintf("commitment to %s (%s) updated %d times but not fulfilled. deadline: %s. status?", commitment.PersonID, commitment.Expectation.Description, len(commitment.History), commitment.Expectation.Deadline.Format("2006-01-02")),
						Required: false,
						Field:    "commitment_status",
					},
				})
			}
		}
	}

	return deviations, nil
}
