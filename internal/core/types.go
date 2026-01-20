package core

import "time"

type Entry struct {
	ID        string
	Timestamp time.Time
	Raw       string
}

type IntentType string

const (
	IntentLog        IntentType = "log"
	IntentProgress   IntentType = "progress"
	IntentCommitment IntentType = "commitment"
	IntentUpdate     IntentType = "update"
	IntentEvent      IntentType = "event"
	IntentCorrection IntentType = "correction"
)

type Intent struct {
	Type       IntentType
	Confidence float64
}

type CommitmentStatus string

const (
	StatusDraft     CommitmentStatus = "draft"
	StatusOpen      CommitmentStatus = "open"
	StatusUpdated   CommitmentStatus = "updated"
	StatusFulfilled CommitmentStatus = "fulfilled"
	StatusViolated  CommitmentStatus = "violated"
	StatusArchived  CommitmentStatus = "archived"
)

type Commitment struct {
	ID           string
	CreatedAt    time.Time
	SourceEntry  string
	PersonID     string
	ProjectID    string
	Expectation  Expectation
	Status       CommitmentStatus
	LastUpdateAt *time.Time
	History      []CommitmentEvent
}

type Expectation struct {
	Description string
	Deadline    time.Time
	Hardness    string // "hard" | "soft"
}

type Person struct {
	ID       string
	Name     string
	Metadata map[string]any
}

type Project struct {
	ID       string
	Priority int
	Metadata map[string]any
}

type CommitmentEvent struct {
	Timestamp   time.Time
	Type        string // "created", "updated", "fulfilled", "violated", "archived"
	Description string
}

type Candidate struct {
	Type       IntentType
	Confidence float64
	Data       map[string]any
	Questions  []Question
}

type Question struct {
	ID       string
	Text     string
	Required bool
	Field    string // field name this question is about
}

type Deviation struct {
	Pattern  PatternType
	Severity string // "low", "medium", "high"
	Question Question
}

type DailyStats struct {
	Date              time.Time
	LogCount          int
	WorkStartTime     *time.Time
	ProgressEntries   int
	CommitmentUpdates int
}

type PatternType string

const (
	PatternLateStart        PatternType = "late_start"
	PatternSparseLogs       PatternType = "sparse_logs"
	PatternCommitmentSilence PatternType = "commitment_silence"
	PatternRepeatedViolations PatternType = "repeated_violations"
	PatternOptimisticStall   PatternType = "optimistic_stall"
)
