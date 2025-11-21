package todo

import (
	"errors"
	"time"

	"github.com/johnmirolha/facienda/internal/recurrence"
)

var (
	ErrEmptyTitle = errors.New("task title cannot be empty")
	ErrNotFound   = errors.New("task not found")
)

type Task struct {
	ID                int64
	Title             string
	Details           string
	Date              time.Time
	Completed         bool
	Skipped           bool
	RecurrencePattern recurrence.Pattern
	Tags              []*Tag
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

func NewTask(title, details string, date time.Time) (*Task, error) {
	if title == "" {
		return nil, ErrEmptyTitle
	}

	now := time.Now()
	return &Task{
		Title:             title,
		Details:           details,
		Date:              date,
		Completed:         false,
		RecurrencePattern: recurrence.PatternNone,
		CreatedAt:         now,
		UpdatedAt:         now,
	}, nil
}

func NewRecurringTask(title, details string, pattern recurrence.Pattern) (*Task, error) {
	if title == "" {
		return nil, ErrEmptyTitle
	}

	// Calculate the first occurrence date
	now := time.Now()
	nextDate, err := pattern.NextOccurrence(now.AddDate(0, 0, -1))
	if err != nil {
		return nil, err
	}

	return &Task{
		Title:             title,
		Details:           details,
		Date:              nextDate,
		Completed:         false,
		RecurrencePattern: pattern,
		CreatedAt:         now,
		UpdatedAt:         now,
	}, nil
}

func (t *Task) Complete() {
	t.Completed = true
	t.UpdatedAt = time.Now()
}

func (t *Task) Incomplete() {
	t.Completed = false
	t.UpdatedAt = time.Now()
}

func (t *Task) Skip() {
	t.Skipped = true
	t.UpdatedAt = time.Now()
}

func (t *Task) Unskip() {
	t.Skipped = false
	t.UpdatedAt = time.Now()
}

func (t *Task) Update(title, details string) error {
	if title == "" {
		return ErrEmptyTitle
	}
	t.Title = title
	t.Details = details
	t.UpdatedAt = time.Now()
	return nil
}

// SetTags updates the task's tags with validation
func (t *Task) SetTags(tags []*Tag) error {
	if err := ValidateTaskTags(tags); err != nil {
		return err
	}
	t.Tags = tags
	t.UpdatedAt = time.Now()
	return nil
}

// GenerateNextInstance creates the next instance of a recurring task
// Returns nil if the task is not recurring
func (t *Task) GenerateNextInstance() (*Task, error) {
	if !t.RecurrencePattern.IsRecurring() {
		return nil, nil
	}

	nextDate, err := t.RecurrencePattern.NextOccurrence(t.Date)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	return &Task{
		Title:             t.Title,
		Details:           t.Details,
		Date:              nextDate,
		Completed:         false,
		RecurrencePattern: t.RecurrencePattern,
		Tags:              t.Tags, // Copy tags to next instance
		CreatedAt:         now,
		UpdatedAt:         now,
	}, nil
}

// IsRecurring returns true if the task has a recurrence pattern
func (t *Task) IsRecurring() bool {
	return t.RecurrencePattern.IsRecurring()
}
