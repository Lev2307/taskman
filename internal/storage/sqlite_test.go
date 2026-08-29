package storage

import (
	"log"
	"path/filepath"
	"slices"
	"testing"
	"time"

	task "github.com/Lev2307/taskman/internal/model"
)

func newTestStore(t *testing.T) *SQLiteStore {
	t.Helper()

	db, err := NewDatabase(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("New database: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return NewSQLiteStore(db)
}

func TestNormalizeTags(t *testing.T) {
	tags := []string{"WOrk", "work", "spoRt", "urgent", "sport", "   Sport"}
	expectedTags := []string{"sport", "urgent", "work"}

	normTags := NormalizeTags(tags)
	if !slices.Equal(expectedTags, normTags) {
		t.Errorf("NormalizeTags() = %q, want %q", normTags, expectedTags)
	}
}

func TestGetAddTask(t *testing.T) {
	db := newTestStore(t)

	normTags := []string{"sport", "urgent"}
	taskToAdd := task.Task{
		Title:     "title test",
		Notes:     "test notes",
		Tags:      []string{"sport", "urgent", "    "},
		CreatedAt: time.Date(2026, 8, 28, 12, 0, 0, 0, time.UTC),
		// DueAt:     time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC),
	}

	addedT, err := db.Add(taskToAdd)
	if err != nil {
		t.Fatalf("add task: %v", err)
	}

	if addedT.ID != 1 {
		t.Fatalf("failed added task data: id = %d, want = %d", addedT.ID, 1)
	}
	if !slices.Equal(addedT.Tags, normTags) {
		t.Fatalf("failed added task data: tags = %q, want = %q", addedT.Tags, normTags)
	}
}

func TestGetTaskByID(t *testing.T) {
	db := newTestStore(t)

	normTags := []string{"sport", "urgent"}
	taskToAdd := task.Task{
		Title:     "title test",
		Notes:     "test notes",
		Tags:      []string{"Urgent", " SPORT ", "sport"},
		CreatedAt: time.Date(2026, 8, 28, 12, 0, 0, 0, time.UTC),
		DueAt:     time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC),
	}

	addedT, err := db.Add(taskToAdd)
	if err != nil {
		t.Fatalf("add task: %v", err)
	}

	foundT, err := db.GetByID(addedT.ID)
	if err != nil {
		t.Fatalf("get task: %v", err)
	}

	if addedT.Title != foundT.Title {
		t.Fatalf("failed get task data: title = %q, want = %q", foundT.Title, addedT.Title)
	}
	if !slices.Equal(foundT.Tags, normTags) {
		t.Fatalf("failed get task data: tags = %q, want = %q", foundT.Tags, normTags)
	}
	if !addedT.DueAt.Equal(foundT.DueAt) {
		t.Fatalf("failed get task data: dueAt = %q, want = %q", foundT.DueAt, addedT.DueAt)
	}
}

func TestTasksList(t *testing.T) {
	db := newTestStore(t)

	normTags := []string{"sport", "urgent"}
	taskToAdd := task.Task{
		Title:     "title test",
		Notes:     "test notes",
		Tags:      []string{"Urgent", " SPORT ", "sport"},
		CreatedAt: time.Date(2026, 8, 28, 12, 0, 0, 0, time.UTC),
		DueAt:     time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC),
	}

	_, err := db.Add(taskToAdd)
	if err != nil {
		t.Fatalf("add task: %v", err)
	}

	tasks, err := db.List()
	if err != nil {
		t.Fatalf("list task: %v", err)
	}

	if !slices.Equal(normTags, tasks[0].Tags) {
		log.Fatalf("failed list task data: tags = %q, want = %q", tasks[0].Tags, normTags)
	}
}

func TestEditTask(t *testing.T) {
	db := newTestStore(t)

	addedT, err := db.Add(task.Task{
		Title:     "old title",
		Notes:     "old notes",
		Tags:      []string{"sport", "urgent"},
		CreatedAt: time.Date(2026, 8, 28, 12, 0, 0, 0, time.UTC),
		DueAt:     time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("add task: %v", err)
	}

	taskToEdit := addedT
	taskToEdit.Title = "new title"
	taskToEdit.Notes = "new notes"
	taskToEdit.Done = true
	taskToEdit.DueAt = time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
	taskToEdit.Tags = []string{"URGENT", " home ", "home"}

	wantTags := []string{"home", "urgent"}

	editedT, err := db.Edit(taskToEdit)
	if err != nil {
		t.Fatalf("edit task: %v", err)
	}

	if editedT.ID != addedT.ID {
		t.Errorf("failed edited task data: id = %d, want = %d", editedT.ID, addedT.ID)
	}
	if editedT.Title != taskToEdit.Title {
		t.Errorf("failed edited task data: title = %q, want = %q", editedT.Title, taskToEdit.Title)
	}
	if editedT.Notes != taskToEdit.Notes {
		t.Errorf("failed edited task data: notes = %q, want = %q", editedT.Notes, taskToEdit.Notes)
	}
	if editedT.Done != taskToEdit.Done {
		t.Errorf("failed edited task data: done = %v, want = %v", editedT.Done, taskToEdit.Done)
	}
	if !editedT.DueAt.Equal(taskToEdit.DueAt) {
		t.Errorf("failed edited task data: dueAt = %q, want = %q", editedT.DueAt, taskToEdit.DueAt)
	}
	// createdAt правится только при создании, Edit его не трогает
	if !editedT.CreatedAt.Equal(addedT.CreatedAt) {
		t.Errorf("failed edited task data: createdAt = %q, want = %q", editedT.CreatedAt, addedT.CreatedAt)
	}
	if !slices.Equal(editedT.Tags, wantTags) {
		t.Errorf("failed edited task data: tags = %q, want = %q", editedT.Tags, wantTags)
	}

	foundT, err := db.GetByID(addedT.ID)
	if err != nil {
		t.Fatalf("get task: %v", err)
	}
	if foundT.Title != taskToEdit.Title {
		t.Errorf("failed stored task data: title = %q, want = %q", foundT.Title, taskToEdit.Title)
	}
	if !slices.Equal(foundT.Tags, wantTags) {
		t.Errorf("failed stored task data: tags = %q, want = %q", foundT.Tags, wantTags)
	}
}
