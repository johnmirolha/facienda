package storage

import (
	"time"

	"github.com/johnmirolha/facienda/internal/todo"
)

type Storage interface {
	// Task operations
	Create(task *todo.Task) error
	GetByID(id int64) (*todo.Task, error)
	List(filter TimeFilter) ([]*todo.Task, error)
	ListByTag(tagName string, filter TimeFilter) ([]*todo.Task, error)
	Update(task *todo.Task) error
	Delete(id int64) error

	// Tag operations
	CreateTag(tag *todo.Tag) error
	GetTagByName(name string) (*todo.Tag, error)
	GetTagByID(id int64) (*todo.Tag, error)
	ListTags() ([]*todo.Tag, error)
	UpdateTag(tag *todo.Tag) error
	DeleteTag(id int64) error

	// Task-Tag associations
	AddTagToTask(taskID int64, tagID int64) error
	RemoveTagFromTask(taskID int64, tagID int64) error
	SetTaskTags(taskID int64, tagIDs []int64) error
	GetTaskTags(taskID int64) ([]*todo.Tag, error)
	GetTasksByTag(tagID int64, filter TimeFilter) ([]*todo.Task, error)
	CountTasksWithTag(tagID int64) (int, error)

	Close() error
}

type TimeFilter int

const (
	FilterAll TimeFilter = iota
	FilterPast
	FilterCurrent
	FilterFuture
)

func StartOfDay(t time.Time) time.Time {
	year, month, day := t.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, t.Location())
}

func EndOfDay(t time.Time) time.Time {
	year, month, day := t.Date()
	return time.Date(year, month, day, 23, 59, 59, 999999999, t.Location())
}
