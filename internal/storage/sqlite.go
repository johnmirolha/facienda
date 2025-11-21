package storage

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/johnmirolha/facienda/internal/recurrence"
	"github.com/johnmirolha/facienda/internal/todo"
	_ "github.com/mattn/go-sqlite3"
)

type SQLiteStorage struct {
	db *sql.DB
}

func NewSQLiteStorage(dbPath string) (*SQLiteStorage, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	s := &SQLiteStorage{db: db}
	if err := s.migrate(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return s, nil
}

func (s *SQLiteStorage) migrate() error {
	query := `
	CREATE TABLE IF NOT EXISTS tasks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		details TEXT,
		date DATETIME NOT NULL,
		completed BOOLEAN NOT NULL DEFAULT 0,
		skipped BOOLEAN NOT NULL DEFAULT 0,
		recurrence_pattern TEXT NOT NULL DEFAULT '',
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL
	);
	CREATE INDEX IF NOT EXISTS idx_tasks_date ON tasks(date);
	CREATE INDEX IF NOT EXISTS idx_tasks_completed ON tasks(completed);
	CREATE INDEX IF NOT EXISTS idx_tasks_skipped ON tasks(skipped);
	`

	if _, err := s.db.Exec(query); err != nil {
		return err
	}

	// Add recurrence_pattern column if it doesn't exist (for existing databases)
	alterQuery := `
	ALTER TABLE tasks ADD COLUMN recurrence_pattern TEXT NOT NULL DEFAULT '';
	`
	// This will fail if the column already exists, which is fine
	s.db.Exec(alterQuery)

	// Add skipped column if it doesn't exist (for existing databases)
	alterSkippedQuery := `
	ALTER TABLE tasks ADD COLUMN skipped BOOLEAN NOT NULL DEFAULT 0;
	`
	// This will fail if the column already exists, which is fine
	s.db.Exec(alterSkippedQuery)

	return nil
}

func (s *SQLiteStorage) Create(task *todo.Task) error {
	query := `
	INSERT INTO tasks (title, details, date, completed, skipped, recurrence_pattern, created_at, updated_at)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := s.db.Exec(query,
		task.Title,
		task.Details,
		task.Date,
		task.Completed,
		task.Skipped,
		string(task.RecurrencePattern),
		task.CreatedAt,
		task.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create task: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	task.ID = id
	return nil
}

func (s *SQLiteStorage) GetByID(id int64) (*todo.Task, error) {
	query := `
	SELECT id, title, details, date, completed, skipped, recurrence_pattern, created_at, updated_at
	FROM tasks
	WHERE id = ?
	`

	task := &todo.Task{}
	var recurrencePattern string
	err := s.db.QueryRow(query, id).Scan(
		&task.ID,
		&task.Title,
		&task.Details,
		&task.Date,
		&task.Completed,
		&task.Skipped,
		&recurrencePattern,
		&task.CreatedAt,
		&task.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, todo.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	task.RecurrencePattern = recurrence.Pattern(recurrencePattern)
	return task, nil
}

func (s *SQLiteStorage) List(filter TimeFilter) ([]*todo.Task, error) {
	query := `
	SELECT id, title, details, date, completed, skipped, recurrence_pattern, created_at, updated_at
	FROM tasks
	WHERE skipped = 0
	`

	args := []interface{}{}
	now := time.Now()
	today := StartOfDay(now)

	switch filter {
	case FilterPast:
		query += " AND date < ?"
		args = append(args, today)
	case FilterCurrent:
		query += " AND date >= ? AND date <= ?"
		args = append(args, today, EndOfDay(now))
	case FilterFuture:
		tomorrow := today.AddDate(0, 0, 1)
		query += " AND date >= ?"
		args = append(args, tomorrow)
	}

	query += " ORDER BY date ASC, created_at ASC"

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list tasks: %w", err)
	}
	defer rows.Close()

	var tasks []*todo.Task
	for rows.Next() {
		task := &todo.Task{}
		var recurrencePattern string
		err := rows.Scan(
			&task.ID,
			&task.Title,
			&task.Details,
			&task.Date,
			&task.Completed,
			&task.Skipped,
			&recurrencePattern,
			&task.CreatedAt,
			&task.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}
		task.RecurrencePattern = recurrence.Pattern(recurrencePattern)
		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating tasks: %w", err)
	}

	return tasks, nil
}

func (s *SQLiteStorage) Update(task *todo.Task) error {
	query := `
	UPDATE tasks
	SET title = ?, details = ?, date = ?, completed = ?, skipped = ?, recurrence_pattern = ?, updated_at = ?
	WHERE id = ?
	`

	result, err := s.db.Exec(query,
		task.Title,
		task.Details,
		task.Date,
		task.Completed,
		task.Skipped,
		string(task.RecurrencePattern),
		task.UpdatedAt,
		task.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return todo.ErrNotFound
	}

	return nil
}

func (s *SQLiteStorage) Delete(id int64) error {
	query := `DELETE FROM tasks WHERE id = ?`

	result, err := s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return todo.ErrNotFound
	}

	return nil
}

func (s *SQLiteStorage) Close() error {
	return s.db.Close()
}
