package stats

import (
	"time"

	"github.com/heywinit/grechen/internal/core"
	"github.com/heywinit/grechen/internal/store"
)

type Stats struct {
	store *store.Store
}

func New(s *store.Store) *Stats {
	return &Stats{store: s}
}

// ComputeDailyStats computes statistics for a given day
func (s *Stats) ComputeDailyStats(date time.Time) (*core.DailyStats, error) {
	// Read daily file
	content, err := s.store.ReadDailyFile(date)
	if err != nil {
		return nil, err
	}

	stats := &core.DailyStats{
		Date: date,
	}

	// Parse content to extract stats
	// Count log entries
	stats.LogCount = countLogEntries(content)

	// Find work start time (first progress entry)
	stats.WorkStartTime = findWorkStartTime(content, date)

	// Count progress entries
	stats.ProgressEntries = countProgressEntries(content)

	// Count commitment updates
	stats.CommitmentUpdates = countCommitmentUpdates(content)

	return stats, nil
}

// ComputeRollingStats computes rolling window statistics
func (s *Stats) ComputeRollingStats(endDate time.Time, days int) (*RollingStats, error) {
	var allStats []*core.DailyStats
	for i := 0; i < days; i++ {
		date := endDate.AddDate(0, 0, -i)
		stats, err := s.ComputeDailyStats(date)
		if err != nil {
			return nil, err
		}
		allStats = append(allStats, stats)
	}

	return aggregateRollingStats(allStats), nil
}

type RollingStats struct {
	AvgLogCount          float64
	AvgWorkStartTime     *time.Time
	AvgProgressEntries   float64
	AvgCommitmentUpdates float64
	Days                 int
}

func aggregateRollingStats(stats []*core.DailyStats) *RollingStats {
	if len(stats) == 0 {
		return &RollingStats{Days: 0}
	}

	rs := &RollingStats{Days: len(stats)}

	var totalLogs, totalProgress, totalUpdates int
	var workStartTimes []time.Time

	for _, s := range stats {
		totalLogs += s.LogCount
		totalProgress += s.ProgressEntries
		totalUpdates += s.CommitmentUpdates
		if s.WorkStartTime != nil {
			workStartTimes = append(workStartTimes, *s.WorkStartTime)
		}
	}

	rs.AvgLogCount = float64(totalLogs) / float64(len(stats))
	rs.AvgProgressEntries = float64(totalProgress) / float64(len(stats))
	rs.AvgCommitmentUpdates = float64(totalUpdates) / float64(len(stats))

	if len(workStartTimes) > 0 {
		// Calculate average time of day
		var totalMinutes int
		for _, t := range workStartTimes {
			totalMinutes += t.Hour()*60 + t.Minute()
		}
		avgMinutes := totalMinutes / len(workStartTimes)
		avgHour := avgMinutes / 60
		avgMin := avgMinutes % 60
		avgTime := time.Date(2000, 1, 1, avgHour, avgMin, 0, 0, time.UTC)
		rs.AvgWorkStartTime = &avgTime
	}

	return rs
}
