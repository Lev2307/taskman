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
