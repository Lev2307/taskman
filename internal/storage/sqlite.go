package storage

import (
	"database/sql"
	"fmt"
	"time"

	task "github.com/Lev2307/taskman/internal/model"
	_ "modernc.org/sqlite"
)

type SQLiteStore struct {
	db *sql.DB
}

var _ task.TaskStore = (*SQLiteStore)(nil)

func NewSQLiteStore(s *sql.DB) *SQLiteStore {
	return &SQLiteStore{db: s}
}

func initDB(db *sql.DB) error {
	const schema = `
	CREATE TABLE IF NOT EXISTS tags (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE
	);
	CREATE TABLE IF NOT EXISTS tasks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		notes TEXT NOT NULL DEFAULT '',
		done BOOLEAN NOT NULL DEFAULT 0,
		createdAt TEXT NOT NULL,
		dueAt TEXT
	);
	CREATE TABLE IF NOT EXISTS task_tags (
		task_id INTEGER NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
		tag_id INTEGER NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
		PRIMARY KEY (task_id, tag_id)
	);
	`
	_, err := db.Exec(schema)
	return err
}

func NewDatabase(dbPath string) (*sql.DB, error) {
	dsn := fmt.Sprintf("file:%s?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&_pragma=foreign_keys(on)", dbPath)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("db connection: %w", err)
	}

	db.SetMaxOpenConns(1)
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping db: %w", err)
	}

	if err := initDB(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("db initialization: %w", err)
	}
	return db, nil
}

func (s *SQLiteStore) List() ([]task.Task, error) {
	rows, err := s.db.Query("SELECT id, title, notes, done, createdAt, dueAt FROM tasks ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []task.Task
	for rows.Next() {
		var t task.Task
		var createdAt string
		var dueAt *string
		if err := rows.Scan(&t.ID, &t.Title, &t.Notes, &t.Done, &createdAt, &dueAt); err != nil {
			return nil, err
		}

		parsedCreatedAt, err := time.Parse(time.RFC3339, createdAt)
		if err != nil {
			return nil, fmt.Errorf("parse createdAt: %w", err)
		}
		t.CreatedAt = parsedCreatedAt

		if dueAt != nil {
			parsedDueAt, err := time.Parse(time.RFC3339, *dueAt)
			if err != nil {
				return nil, fmt.Errorf("parse dueAt: %w", err)
			}
			t.DueAt = parsedDueAt
		}

		tasks = append(tasks, t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return tasks, nil
}

func (s *SQLiteStore) GetByID(taskID int) (task.Task, error) {
	return task.Task{}, nil
}

func (s *SQLiteStore) Add(t task.Task) (task.Task, error) {
	return task.Task{}, nil
}

func (s *SQLiteStore) Edit(t task.Task) (task.Task, error) {
	return task.Task{}, nil
}

func (s *SQLiteStore) Delete(taskID int) error {
	return nil
}

func (s *SQLiteStore) SetDone(taskID int, done bool) (task.Task, error) {
	return task.Task{}, nil
}
