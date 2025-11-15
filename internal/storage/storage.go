package storage

import (
	"time"

	"github.com/johnmirolha/facienda/internal/todo"
)

type Storage interface {
	Create(task *todo.Task) error
	GetByID(id int64) (*todo.Task, error)
	List(filter TimeFilter) ([]*todo.Task, error)
	Update(task *todo.Task) error
	Delete(id int64) error
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
