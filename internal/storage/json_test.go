package storage

import (
	"fmt"
	"path/filepath"
	"sync"
	"testing"

	task "github.com/Lev2307/taskman/internal/model"
)

// 50 горутин одновременно добавляют по задаче. Все 50 должны доехать до файла,
// и все ID должны быть разными.
func TestAddConcurrent(t *testing.T) {
	s := NewStore(filepath.Join(t.TempDir(), "tasks.json"))

	const n = 50
	var wg sync.WaitGroup
	for i := range n {
		wg.Go(func() {
			if _, err := s.Add(task.Task{Title: fmt.Sprintf("task %d", i)}); err != nil {
				t.Errorf("Add: %v", err)
			}
		})
	}
	wg.Wait()

	tasks, err := s.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}

	if len(tasks) != n {
		t.Errorf("задач в файле: %d, ожидалось %d — записи потерялись", len(tasks), n)
	}

	seen := make(map[int]bool, len(tasks))
	for _, tk := range tasks {
		if seen[tk.ID] {
			t.Errorf("ID %d выдан больше одного раза", tk.ID)
		}
		seen[tk.ID] = true
	}
}
