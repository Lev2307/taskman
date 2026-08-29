package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"slices"
	"strings"
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

func NormalizeTags(tags []string) []string {
	nonEmptyTags := make([]string, 0, len(tags))
	for i := range tags {
		tag := strings.ToLower(tags[i])
		tag = strings.TrimSpace(tag)
		if tag != "" {
			nonEmptyTags = append(nonEmptyTags, tag)
		}
	}
	tagNames := make(map[string]struct{}, len(nonEmptyTags))
	for i := range nonEmptyTags {
		if _, ok := tagNames[nonEmptyTags[i]]; !ok {
			tagNames[nonEmptyTags[i]] = struct{}{}
		}
	}
	newTags := make([]string, 0, len(nonEmptyTags))
	for i := range tagNames {
		newTags = append(newTags, i)
	}

	slices.Sort(newTags)
	return newTags
}

func upsertTags(tx *sql.Tx, t task.Task, dataTags []string) error {
	tagsNorm := NormalizeTags(dataTags)
	for _, tag := range tagsNorm {
		tagQuery := `
		INSERT INTO tags (name) VALUES (?)
		ON CONFLICT(name) DO UPDATE SET name = excluded.name
		RETURNING id;
		`
		var tagID int64
		err := tx.QueryRow(tagQuery, tag).Scan(&tagID)
		if err != nil {
			return err
		}
		if _, err := tx.Exec("INSERT OR IGNORE INTO task_tags (task_id, tag_id) VALUES (?, ?)", t.ID, tagID); err != nil {
			return err
		}
	}
	return nil
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
	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	rows, err := tx.Query("SELECT id, title, notes, done, createdAt, dueAt FROM tasks ORDER BY id")
	if err != nil {
		return nil, err
	}

	var tasks []task.Task
	for rows.Next() {
		t, err := scanTask(rows)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	defer rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}

	taskMap := make(map[int][]string, len(tasks))
	for i := range tasks {
		_, ok := taskMap[tasks[i].ID]
		if !ok {
			tagQuery := `
			SELECT tags.name from tags
			JOIN task_tags tt ON tt.tag_id = tags.id WHERE tt.task_id = ? ORDER BY tags.name;
			`
			tagRows, err := tx.Query(tagQuery, tasks[i].ID)
			if err != nil {
				return nil, err
			}

			var taskTags []string
			for tagRows.Next() {
				var tagName string
				if err := tagRows.Scan(&tagName); err != nil {
					tagRows.Close()
					return nil, err
				}
				taskTags = append(taskTags, tagName)
			}

			if err := tagRows.Err(); err != nil {
				return nil, err
			}

			taskMap[tasks[i].ID] = taskTags
		}
	}

	for i := range tasks {
		if tTags, ok := taskMap[tasks[i].ID]; ok {
			tasks[i].Tags = tTags
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return tasks, nil
}

func (s *SQLiteStore) GetByID(taskID int) (task.Task, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return task.Task{}, err
	}
	defer tx.Rollback()
	row := tx.QueryRow("SELECT id, title, notes, done, createdAt, dueAt FROM tasks WHERE id = ?;", taskID)

	t, err := scanTask(row)
	if err != nil {
		return task.Task{}, err
	}

	tagsQuery := `
	SELECT tag.name FROM tags tag 
	JOIN task_tags tt ON tt.tag_id = tag.id WHERE tt.task_id = ? ORDER BY tag.name;
	`
	rows, err := tx.Query(tagsQuery, taskID)
	if err != nil {
		return task.Task{}, err
	}
	defer rows.Close()

	var tags []string
	for rows.Next() {
		var tag string
		if err := rows.Scan(&tag); err != nil {
			return task.Task{}, err
		}
		tags = append(tags, tag)
	}
	if err := rows.Err(); err != nil {
		return task.Task{}, err
	}

	if err := tx.Commit(); err != nil {
		return task.Task{}, err
	}
	t.Tags = tags
	return t, nil
}

func (s *SQLiteStore) Add(t task.Task) (task.Task, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return task.Task{}, err
	}
	defer tx.Rollback()

	createdAt := t.CreatedAt.UTC().Format(time.RFC3339)
	var dueAt *string
	if !t.DueAt.IsZero() {
		dueAtFormatted := t.DueAt.UTC().Format(time.RFC3339)
		dueAt = &dueAtFormatted
	}
	query := `
	INSERT INTO tasks (title, notes, done, createdAt, dueAt)
	VALUES (?, ?, ?, ?, ?)
	RETURNING id, title, notes, done, createdAt, dueAt;
	`
	row := tx.QueryRow(query, t.Title, t.Notes, t.Done, createdAt, dueAt)
	tScanned, err := scanTask(row)
	if err != nil {
		return task.Task{}, err
	}

	tagsNorm := NormalizeTags(t.Tags)
	if err := upsertTags(tx, tScanned, tagsNorm); err != nil {
		return task.Task{}, err
	}

	if err := tx.Commit(); err != nil {
		return task.Task{}, err
	}
	tScanned.Tags = tagsNorm
	return tScanned, nil
}

func (s *SQLiteStore) Edit(t task.Task) (task.Task, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return task.Task{}, err
	}
	defer tx.Rollback()
	var dueAt *string
	if !t.DueAt.IsZero() {
		dueAtFormatted := t.DueAt.UTC().Format(time.RFC3339)
		dueAt = &dueAtFormatted
	}
	query := `
	UPDATE tasks
	SET title = ?, notes = ?, done = ?, dueAt = ?
	WHERE id = ?
	RETURNING id, title, notes, done, createdAt, dueAt;
	`
	row := tx.QueryRow(query, t.Title, t.Notes, t.Done, dueAt, t.ID)

	taskScanned, err := scanTask(row)
	if err != nil {
		return task.Task{}, err
	}

	deletePrevTaskConsQuery := `DELETE FROM task_tags WHERE task_tags.task_id = ?`
	res, err := tx.Exec(deletePrevTaskConsQuery, taskScanned.ID)
	if err != nil {
		return task.Task{}, err
	}

	n, err := res.RowsAffected()
	if n == 0 {
		return task.Task{}, err
	}

	tagsNorm := NormalizeTags(t.Tags)
	if err := upsertTags(tx, taskScanned, tagsNorm); err != nil {
		return task.Task{}, err
	}

	if err := tx.Commit(); err != nil {
		return task.Task{}, err
	}
	taskScanned.Tags = tagsNorm

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
	tx, err := s.db.Begin()
	if err != nil {
		return task.Task{}, err
	}
	defer tx.Rollback()
	query := `
	UPDATE tasks
	SET done = ?
	WHERE id = ?
	RETURNING id, title, notes, done, createdAt, dueAt;
	`

	row := tx.QueryRow(query, done, taskID)
	scannedT, err := scanTask(row)
	if err != nil {
		return task.Task{}, err
	}

	tagsQuery := `
	SELECT tag.name FROM tags tag 
	JOIN task_tags tt ON tt.tag_id = tag.id WHERE tt.task_id = ? ORDER BY tag.name;
	`
	rows, err := tx.Query(tagsQuery, taskID)
	if err != nil {
		return task.Task{}, err
	}
	defer rows.Close()

	var tags []string
	for rows.Next() {
		var tag string
		if err := rows.Scan(&tag); err != nil {
			return task.Task{}, err
		}
		tags = append(tags, tag)
	}

	if err := rows.Err(); err != nil {
		return task.Task{}, err
	}

	if err := tx.Commit(); err != nil {
		return task.Task{}, err
	}
	scannedT.Tags = tags
	return scannedT, nil
}
