package storage

import (
	"os"
	"testing"
	"time"

	"github.com/johnmirolha/facienda/internal/todo"
)

func setupTestDB(t *testing.T) (*SQLiteStorage, func()) {
	t.Helper()

	tmpFile, err := os.CreateTemp("", "facienda_test_*.db")
	if err != nil {
		t.Fatalf("failed to create temp db: %v", err)
	}
	tmpFile.Close()

	store, err := NewSQLiteStorage(tmpFile.Name())
	if err != nil {
		os.Remove(tmpFile.Name())
		t.Fatalf("failed to create storage: %v", err)
	}

	cleanup := func() {
		store.Close()
		os.Remove(tmpFile.Name())
	}

	return store, cleanup
}

func TestIntegration_TaskLifecycle(t *testing.T) {
	store, cleanup := setupTestDB(t)
	defer cleanup()

	task, err := todo.NewTask("Buy groceries", "Milk, eggs, bread", time.Now())
	if err != nil {
		t.Fatalf("failed to create task: %v", err)
	}

	if err := store.Create(task); err != nil {
		t.Fatalf("failed to create task in db: %v", err)
	}

	if task.ID == 0 {
		t.Error("expected task ID to be set")
	}

	retrieved, err := store.GetByID(task.ID)
	if err != nil {
		t.Fatalf("failed to get task: %v", err)
	}

	if retrieved.Title != task.Title {
		t.Errorf("title mismatch: got %q, want %q", retrieved.Title, task.Title)
	}
	if retrieved.Details != task.Details {
		t.Errorf("details mismatch: got %q, want %q", retrieved.Details, task.Details)
	}
	if retrieved.Completed {
		t.Error("expected task to be incomplete")
	}

	retrieved.Complete()
	if err := store.Update(retrieved); err != nil {
		t.Fatalf("failed to update task: %v", err)
	}

	updated, err := store.GetByID(task.ID)
	if err != nil {
		t.Fatalf("failed to get updated task: %v", err)
	}
	if !updated.Completed {
		t.Error("expected task to be completed")
	}

	if err := store.Delete(task.ID); err != nil {
		t.Fatalf("failed to delete task: %v", err)
	}

	_, err = store.GetByID(task.ID)
	if err != todo.ErrNotFound {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}
}

func TestIntegration_TimeFilters(t *testing.T) {
	store, cleanup := setupTestDB(t)
	defer cleanup()

	now := time.Now()
	yesterday := now.AddDate(0, 0, -1)
	tomorrow := now.AddDate(0, 0, 1)

	tasks := []*todo.Task{
		{Title: "Past task", Details: "", Date: yesterday, Completed: false, CreatedAt: now, UpdatedAt: now},
		{Title: "Current task", Details: "", Date: now, Completed: false, CreatedAt: now, UpdatedAt: now},
		{Title: "Future task", Details: "", Date: tomorrow, Completed: false, CreatedAt: now, UpdatedAt: now},
	}

	for _, task := range tasks {
		if err := store.Create(task); err != nil {
			t.Fatalf("failed to create task: %v", err)
		}
	}

	pastTasks, err := store.List(FilterPast)
	if err != nil {
		t.Fatalf("failed to list past tasks: %v", err)
	}
	if len(pastTasks) != 1 {
		t.Errorf("expected 1 past task, got %d", len(pastTasks))
	}
	if len(pastTasks) > 0 && pastTasks[0].Title != "Past task" {
		t.Errorf("expected 'Past task', got %q", pastTasks[0].Title)
	}

	currentTasks, err := store.List(FilterCurrent)
	if err != nil {
		t.Fatalf("failed to list current tasks: %v", err)
	}
	if len(currentTasks) != 1 {
		t.Errorf("expected 1 current task, got %d", len(currentTasks))
	}
	if len(currentTasks) > 0 && currentTasks[0].Title != "Current task" {
		t.Errorf("expected 'Current task', got %q", currentTasks[0].Title)
	}

	futureTasks, err := store.List(FilterFuture)
	if err != nil {
		t.Fatalf("failed to list future tasks: %v", err)
	}
	if len(futureTasks) != 1 {
		t.Errorf("expected 1 future task, got %d", len(futureTasks))
	}
	if len(futureTasks) > 0 && futureTasks[0].Title != "Future task" {
		t.Errorf("expected 'Future task', got %q", futureTasks[0].Title)
	}

	allTasks, err := store.List(FilterAll)
	if err != nil {
		t.Fatalf("failed to list all tasks: %v", err)
	}
	if len(allTasks) != 3 {
		t.Errorf("expected 3 tasks total, got %d", len(allTasks))
	}
}

func TestIntegration_EditTask(t *testing.T) {
	store, cleanup := setupTestDB(t)
	defer cleanup()

	task, _ := todo.NewTask("Original title", "Original details", time.Now())
	if err := store.Create(task); err != nil {
		t.Fatalf("failed to create task: %v", err)
	}

	if err := task.Update("Updated title", "Updated details"); err != nil {
		t.Fatalf("failed to update task: %v", err)
	}

	if err := store.Update(task); err != nil {
		t.Fatalf("failed to save updated task: %v", err)
	}

	retrieved, err := store.GetByID(task.ID)
	if err != nil {
		t.Fatalf("failed to get task: %v", err)
	}

	if retrieved.Title != "Updated title" {
		t.Errorf("title not updated: got %q, want %q", retrieved.Title, "Updated title")
	}
	if retrieved.Details != "Updated details" {
		t.Errorf("details not updated: got %q, want %q", retrieved.Details, "Updated details")
	}
}

func TestIntegration_CompleteIncomplete(t *testing.T) {
	store, cleanup := setupTestDB(t)
	defer cleanup()

	task, _ := todo.NewTask("Test task", "", time.Now())
	if err := store.Create(task); err != nil {
		t.Fatalf("failed to create task: %v", err)
	}

	task.Complete()
	if err := store.Update(task); err != nil {
		t.Fatalf("failed to complete task: %v", err)
	}

	retrieved, _ := store.GetByID(task.ID)
	if !retrieved.Completed {
		t.Error("expected task to be completed")
	}

	retrieved.Incomplete()
	if err := store.Update(retrieved); err != nil {
		t.Fatalf("failed to mark incomplete: %v", err)
	}

	retrieved, _ = store.GetByID(task.ID)
	if retrieved.Completed {
		t.Error("expected task to be incomplete")
	}
}

func TestIntegration_SkipUnskip(t *testing.T) {
	store, cleanup := setupTestDB(t)
	defer cleanup()

	task, _ := todo.NewTask("Test task", "", time.Now())
	if err := store.Create(task); err != nil {
		t.Fatalf("failed to create task: %v", err)
	}

	// Verify task is not skipped initially
	if task.Skipped {
		t.Error("expected task to not be skipped initially")
	}

	// Skip the task
	task.Skip()
	if err := store.Update(task); err != nil {
		t.Fatalf("failed to skip task: %v", err)
	}

	retrieved, _ := store.GetByID(task.ID)
	if !retrieved.Skipped {
		t.Error("expected task to be skipped")
	}

	// Unskip the task
	retrieved.Unskip()
	if err := store.Update(retrieved); err != nil {
		t.Fatalf("failed to unskip task: %v", err)
	}

	retrieved, _ = store.GetByID(task.ID)
	if retrieved.Skipped {
		t.Error("expected task to be unskipped")
	}
}

func TestIntegration_SkippedTasksNotInList(t *testing.T) {
	store, cleanup := setupTestDB(t)
	defer cleanup()

	now := time.Now()

	// Create two tasks
	task1, _ := todo.NewTask("Task 1", "", now)
	task2, _ := todo.NewTask("Task 2", "", now)

	if err := store.Create(task1); err != nil {
		t.Fatalf("failed to create task1: %v", err)
	}
	if err := store.Create(task2); err != nil {
		t.Fatalf("failed to create task2: %v", err)
	}

	// List should show both tasks
	tasks, err := store.List(FilterCurrent)
	if err != nil {
		t.Fatalf("failed to list tasks: %v", err)
	}
	if len(tasks) != 2 {
		t.Errorf("expected 2 tasks, got %d", len(tasks))
	}

	// Skip task1
	task1.Skip()
	if err := store.Update(task1); err != nil {
		t.Fatalf("failed to skip task1: %v", err)
	}

	// List should now show only task2
	tasks, err = store.List(FilterCurrent)
	if err != nil {
		t.Fatalf("failed to list tasks: %v", err)
	}
	if len(tasks) != 1 {
		t.Errorf("expected 1 task after skip, got %d", len(tasks))
	}
	if len(tasks) > 0 && tasks[0].Title != "Task 2" {
		t.Errorf("expected 'Task 2', got %q", tasks[0].Title)
	}

	// Unskip task1
	task1.Unskip()
	if err := store.Update(task1); err != nil {
		t.Fatalf("failed to unskip task1: %v", err)
	}

	// List should show both tasks again
	tasks, err = store.List(FilterCurrent)
	if err != nil {
		t.Fatalf("failed to list tasks: %v", err)
	}
	if len(tasks) != 2 {
		t.Errorf("expected 2 tasks after unskip, got %d", len(tasks))
	}
}
