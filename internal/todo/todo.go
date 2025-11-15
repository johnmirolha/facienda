package todo

import (
	"errors"
	"time"
)

var (
	ErrEmptyTitle = errors.New("task title cannot be empty")
	ErrNotFound   = errors.New("task not found")
)

type Task struct {
	ID        int64
	Title     string
	Details   string
	Date      time.Time
	Completed bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewTask(title, details string, date time.Time) (*Task, error) {
	if title == "" {
		return nil, ErrEmptyTitle
	}

	now := time.Now()
	return &Task{
		Title:     title,
		Details:   details,
		Date:      date,
		Completed: false,
		CreatedAt: now,
		UpdatedAt: now,
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

func (t *Task) Update(title, details string) error {
	if title == "" {
		return ErrEmptyTitle
	}
	t.Title = title
	t.Details = details
	t.UpdatedAt = time.Now()
	return nil
}
