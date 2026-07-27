package task

import "time"

type Task struct {
	ID        int
	Title     string
	Notes     string
	Done      bool
	Tags      []string
	CreatedAt time.Time
	DueAt     time.Time
}
