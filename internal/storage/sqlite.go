package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	task "github.com/Lev2307/taskman/internal/model"
	_ "modernc.org/sqlite"
)

type SQLiteStore struct {
	db *sql.DB
}

var _ task.TaskStore = (*SQLiteStore)(nil)

type rowScanner interface {
	Scan(dest ...any) error
}

func scanTask(sc rowScanner) (task.Task, error) {
	var t task.Task
	var createdAt string
	var dueAt *string
	if err := sc.Scan(&t.ID, &t.Title, &t.Notes, &t.Done, &createdAt, &dueAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return task.Task{}, ErrTaskNotFound
		}
		return task.Task{}, err
	}

	parsedCreatedAt, err := time.Parse(time.RFC3339, createdAt)
	if err != nil {
		return task.Task{}, fmt.Errorf("parse createdAt: %w", err)
	}
	t.CreatedAt = parsedCreatedAt

	if dueAt != nil {
		parsedDueAt, err := time.Parse(time.RFC3339, *dueAt)
		if err != nil {
			return task.Task{}, fmt.Errorf("parse dueAt: %w", err)
		}
		t.DueAt = parsedDueAt
	}

	return t, nil
}

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
	rows, err := s.db.Query("SELECT id, title, notes, done, createdAt, dueAt FROM tasks ORDER BY id;")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []task.Task
	for rows.Next() {
		t, err := scanTask(rows)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return tasks, nil
}

func (s *SQLiteStore) GetByID(taskID int) (task.Task, error) {
	row := s.db.QueryRow("SELECT id, title, notes, done, createdAt, dueAt FROM tasks WHERE id = ?;", taskID)

	t, err := scanTask(row)
	if err != nil {
		return task.Task{}, err
	}
	return t, nil
}

func (s *SQLiteStore) Add(t task.Task) (task.Task, error) {
	createdAt := t.CreatedAt.UTC().Format(time.RFC3339)
	dueAt := t.DueAt.UTC().Format(time.RFC3339)
	query := `
	INSERT INTO tasks (title, notes, done, createdAt, dueAt)
	VALUES (?, ?, ?, ?, ?)
	RETURNING id, title, notes, done, createdAt, dueAt;
	`
	row := s.db.QueryRow(query, t.Title, t.Notes, t.Done, createdAt, dueAt)

	tScanned, err := scanTask(row)
	if err != nil {
		return task.Task{}, err
	}
	return tScanned, nil
}

func (s *SQLiteStore) Edit(t task.Task) (task.Task, error) {
	dueAt := t.DueAt.UTC().Format(time.RFC3339)
	query := `
	UPDATE tasks
	SET title = ?, notes = ?, done = ?, dueAt = ?
	WHERE id = ?
	RETURNING id, title, notes, done, createdAt, dueAt;
	`
	row := s.db.QueryRow(query, t.Title, t.Notes, t.Done, dueAt, t.ID)

	taskScanned, err := scanTask(row)
	if err != nil {
		return task.Task{}, err
	}
	return taskScanned, nil
}

func (s *SQLiteStore) Delete(taskID int) error {
	query := `DELETE FROM tasks WHERE id = ?;`

	res, err := s.db.Exec(query, taskID)
	if err != nil {
		return err
	}

	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrTaskNotFound
	}
	return nil
}

func (s *SQLiteStore) SetDone(taskID int, done bool) (task.Task, error) {
	query := `
	UPDATE tasks
	SET done = ?
	WHERE id = ?
	RETURNING id, title, notes, done, createdAt, dueAt;
	`

	row := s.db.QueryRow(query, done, taskID)

	scannedT, err := scanTask(row)
	if err != nil {
		return task.Task{}, err
	}
	return scannedT, nil
}
