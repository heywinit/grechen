package patterns

import (
	"time"

	"github.com/heywinit/grechen/internal/core"
	"github.com/heywinit/grechen/internal/store"
	"github.com/heywinit/grechen/internal/stats"
)

type Patterns struct {
	store *store.Store
	stats *stats.Stats
}

func New(s *store.Store, st *stats.Stats) *Patterns {
	return &Patterns{
		store: s,
		stats: st,
	}
}

// Evaluate evaluates patterns for a given date and returns deviations
func (p *Patterns) Evaluate(date time.Time, rollingStats *stats.RollingStats) ([]core.Deviation, error) {
	var deviations []core.Deviation

	// Get today's stats
	todayStats, err := p.stats.ComputeDailyStats(date)
	if err != nil {
		return nil, err
	}

	// Check each pattern
	if dev := p.detectLateStart(todayStats, rollingStats); dev != nil {
		deviations = append(deviations, *dev)
	}

	if dev := p.detectSparseLogs(todayStats, rollingStats); dev != nil {
		deviations = append(deviations, *dev)
	}

	commitmentDeviations, err := p.detectCommitmentSilence(date)
	if err != nil {
		return nil, err
	}
	deviations = append(deviations, commitmentDeviations...)

	violationDeviations, err := p.detectRepeatedViolations()
	if err != nil {
		return nil, err
	}
	deviations = append(deviations, violationDeviations...)

	stallDeviations, err := p.detectOptimisticStall()
	if err != nil {
		return nil, err
	}
	deviations = append(deviations, stallDeviations...)

	return deviations, nil
}
