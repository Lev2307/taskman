package task

import "time"

type Task struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Notes     string    `json:"notes"`
	Done      bool      `json:"done"`
	Tags      []string  `json:"tags"`
	CreatedAt time.Time `json:"createdAt"`
	DueAt     time.Time `json:"dueAt"`
}

type TaskStore interface {
	List() ([]Task, error)
	GetByID(taskID int) (Task, error)
	Add(t Task) (Task, error)
	Edit(t Task) (Task, error)
	Delete(taskID int) error
	SetDone(taskID int, done bool) (Task, error)
}
