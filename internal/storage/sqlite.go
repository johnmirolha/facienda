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

	CREATE TABLE IF NOT EXISTS tags (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT UNIQUE NOT NULL,
		created_at DATETIME NOT NULL
	);
	CREATE INDEX IF NOT EXISTS idx_tags_name ON tags(name);

	CREATE TABLE IF NOT EXISTS task_tags (
		task_id INTEGER NOT NULL,
		tag_id INTEGER NOT NULL,
		PRIMARY KEY (task_id, tag_id),
		FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE,
		FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE
	);
	CREATE INDEX IF NOT EXISTS idx_task_tags_task_id ON task_tags(task_id);
	CREATE INDEX IF NOT EXISTS idx_task_tags_tag_id ON task_tags(tag_id);
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
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	query := `
	INSERT INTO tasks (title, details, date, completed, skipped, recurrence_pattern, created_at, updated_at)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := tx.Exec(query,
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

	// Associate tags with the task
	if len(task.Tags) > 0 {
		tagIDs := make([]int64, len(task.Tags))
		for i, tag := range task.Tags {
			tagIDs[i] = tag.ID
		}
		if err := s.setTaskTagsInTx(tx, id, tagIDs); err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

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

	// Load tags for the task
	tags, err := s.GetTaskTags(id)
	if err != nil {
		return nil, err
	}
	task.Tags = tags

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

	// Load tags for all tasks
	for _, task := range tasks {
		tags, err := s.GetTaskTags(task.ID)
		if err != nil {
			return nil, err
		}
		task.Tags = tags
	}

	return tasks, nil
}

func (s *SQLiteStorage) Update(task *todo.Task) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	query := `
	UPDATE tasks
	SET title = ?, details = ?, date = ?, completed = ?, skipped = ?, recurrence_pattern = ?, updated_at = ?
	WHERE id = ?
	`

	result, err := tx.Exec(query,
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

	// Update tags
	tagIDs := make([]int64, len(task.Tags))
	for i, tag := range task.Tags {
		tagIDs[i] = tag.ID
	}
	if err := s.setTaskTagsInTx(tx, task.ID, tagIDs); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
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

// Tag CRUD operations

func (s *SQLiteStorage) CreateTag(tag *todo.Tag) error {
	query := `INSERT INTO tags (name, created_at) VALUES (?, ?)`

	result, err := s.db.Exec(query, tag.Name, tag.CreatedAt)
	if err != nil {
		// Check for unique constraint violation
		if err.Error() == "UNIQUE constraint failed: tags.name" {
			return todo.ErrTagAlreadyExists
		}
		return fmt.Errorf("failed to create tag: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	tag.ID = id
	return nil
}

func (s *SQLiteStorage) GetTagByName(name string) (*todo.Tag, error) {
	query := `SELECT id, name, created_at FROM tags WHERE name = ?`

	tag := &todo.Tag{}
	err := s.db.QueryRow(query, name).Scan(&tag.ID, &tag.Name, &tag.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, todo.ErrTagNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get tag: %w", err)
	}

	return tag, nil
}

func (s *SQLiteStorage) GetTagByID(id int64) (*todo.Tag, error) {
	query := `SELECT id, name, created_at FROM tags WHERE id = ?`

	tag := &todo.Tag{}
	err := s.db.QueryRow(query, id).Scan(&tag.ID, &tag.Name, &tag.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, todo.ErrTagNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get tag: %w", err)
	}

	return tag, nil
}

func (s *SQLiteStorage) ListTags() ([]*todo.Tag, error) {
	query := `SELECT id, name, created_at FROM tags ORDER BY name ASC`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list tags: %w", err)
	}
	defer rows.Close()

	var tags []*todo.Tag
	for rows.Next() {
		tag := &todo.Tag{}
		if err := rows.Scan(&tag.ID, &tag.Name, &tag.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan tag: %w", err)
		}
		tags = append(tags, tag)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating tags: %w", err)
	}

	return tags, nil
}

func (s *SQLiteStorage) UpdateTag(tag *todo.Tag) error {
	query := `UPDATE tags SET name = ? WHERE id = ?`

	result, err := s.db.Exec(query, tag.Name, tag.ID)
	if err != nil {
		// Check for unique constraint violation
		if err.Error() == "UNIQUE constraint failed: tags.name" {
			return todo.ErrTagAlreadyExists
		}
		return fmt.Errorf("failed to update tag: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return todo.ErrTagNotFound
	}

	return nil
}

func (s *SQLiteStorage) DeleteTag(id int64) error {
	// First check if the tag is in use
	count, err := s.CountTasksWithTag(id)
	if err != nil {
		return err
	}
	if count > 0 {
		return todo.ErrTagInUse
	}

	query := `DELETE FROM tags WHERE id = ?`

	result, err := s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete tag: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return todo.ErrTagNotFound
	}

	return nil
}

// Task-Tag association operations

func (s *SQLiteStorage) AddTagToTask(taskID int64, tagID int64) error {
	query := `INSERT INTO task_tags (task_id, tag_id) VALUES (?, ?)`

	_, err := s.db.Exec(query, taskID, tagID)
	if err != nil {
		return fmt.Errorf("failed to add tag to task: %w", err)
	}

	return nil
}

func (s *SQLiteStorage) RemoveTagFromTask(taskID int64, tagID int64) error {
	query := `DELETE FROM task_tags WHERE task_id = ? AND tag_id = ?`

	_, err := s.db.Exec(query, taskID, tagID)
	if err != nil {
		return fmt.Errorf("failed to remove tag from task: %w", err)
	}

	return nil
}

func (s *SQLiteStorage) SetTaskTags(taskID int64, tagIDs []int64) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	if err := s.setTaskTagsInTx(tx, taskID, tagIDs); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// setTaskTagsInTx is a helper function to set task tags within a transaction
func (s *SQLiteStorage) setTaskTagsInTx(tx *sql.Tx, taskID int64, tagIDs []int64) error {
	// First, delete all existing tags for this task
	deleteQuery := `DELETE FROM task_tags WHERE task_id = ?`
	if _, err := tx.Exec(deleteQuery, taskID); err != nil {
		return fmt.Errorf("failed to delete existing tags: %w", err)
	}

	// Then, insert the new tags
	if len(tagIDs) > 0 {
		insertQuery := `INSERT INTO task_tags (task_id, tag_id) VALUES (?, ?)`
		for _, tagID := range tagIDs {
			if _, err := tx.Exec(insertQuery, taskID, tagID); err != nil {
				return fmt.Errorf("failed to insert tag: %w", err)
			}
		}
	}

	return nil
}

func (s *SQLiteStorage) GetTaskTags(taskID int64) ([]*todo.Tag, error) {
	query := `
	SELECT t.id, t.name, t.created_at
	FROM tags t
	INNER JOIN task_tags tt ON t.id = tt.tag_id
	WHERE tt.task_id = ?
	ORDER BY t.name ASC
	`

	rows, err := s.db.Query(query, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get task tags: %w", err)
	}
	defer rows.Close()

	var tags []*todo.Tag
	for rows.Next() {
		tag := &todo.Tag{}
		if err := rows.Scan(&tag.ID, &tag.Name, &tag.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan tag: %w", err)
		}
		tags = append(tags, tag)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating tags: %w", err)
	}

	return tags, nil
}

func (s *SQLiteStorage) GetTasksByTag(tagID int64, filter TimeFilter) ([]*todo.Task, error) {
	query := `
	SELECT t.id, t.title, t.details, t.date, t.completed, t.skipped, t.recurrence_pattern, t.created_at, t.updated_at
	FROM tasks t
	INNER JOIN task_tags tt ON t.id = tt.task_id
	WHERE tt.tag_id = ? AND t.skipped = 0
	`

	args := []interface{}{tagID}
	now := time.Now()
	today := StartOfDay(now)

	switch filter {
	case FilterPast:
		query += " AND t.date < ?"
		args = append(args, today)
	case FilterCurrent:
		query += " AND t.date >= ? AND t.date <= ?"
		args = append(args, today, EndOfDay(now))
	case FilterFuture:
		tomorrow := today.AddDate(0, 0, 1)
		query += " AND t.date >= ?"
		args = append(args, tomorrow)
	}

	query += " ORDER BY t.date ASC, t.created_at ASC"

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get tasks by tag: %w", err)
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

	// Load tags for all tasks
	for _, task := range tasks {
		tags, err := s.GetTaskTags(task.ID)
		if err != nil {
			return nil, err
		}
		task.Tags = tags
	}

	return tasks, nil
}

func (s *SQLiteStorage) ListByTag(tagName string, filter TimeFilter) ([]*todo.Task, error) {
	tag, err := s.GetTagByName(tagName)
	if err != nil {
		return nil, err
	}

	return s.GetTasksByTag(tag.ID, filter)
}

func (s *SQLiteStorage) CountTasksWithTag(tagID int64) (int, error) {
	query := `SELECT COUNT(*) FROM task_tags WHERE tag_id = ?`

	var count int
	err := s.db.QueryRow(query, tagID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count tasks with tag: %w", err)
	}

	return count, nil
}
