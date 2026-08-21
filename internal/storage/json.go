package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"

	task "github.com/Lev2307/taskman/internal/model"
)

var ErrTaskNotFound = errors.New("task with given id not found")

type Store struct {
	mu   sync.Mutex
	path string
}

func NewStore(path string) *Store {
	return &Store{path: path}
}

// load чтение файла. Вызывающий обязан держать s.mu.
func (s *Store) load() ([]task.Task, error) {
	f, err := os.Open(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("open file %s: %w", s.path, err)
	}
	defer f.Close()

	var tasks []task.Task
	if err := json.NewDecoder(f).Decode(&tasks); err != nil {
		if errors.Is(err, io.EOF) {
			return nil, nil
		}
		return nil, fmt.Errorf("decode: %w", err)
	}
	return tasks, nil
}

func (s *Store) save(tasks []task.Task) error {
	f, err := os.Create(s.path)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "   ")
	if err := enc.Encode(tasks); err != nil {
		return fmt.Errorf("encode: %w", err)
	}
	return nil
}

// при помощи mutex
func (s *Store) Add(t task.Task) (task.Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	tasks, err := s.load()
	if err != nil {
		return task.Task{}, err
	}
	maxID := 0
	for _, cur := range tasks {
		if cur.ID > maxID {
			maxID = cur.ID
		}
	}
	t.ID = maxID + 1
	tasks = append(tasks, t)
	if err := s.save(tasks); err != nil {
		return task.Task{}, err
	}
	return t, nil
}

func (s *Store) GetByID(taskID int) (task.Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	tasks, err := s.load()
	if err != nil {
		return task.Task{}, err
	}
	tasksMap := make(map[int]task.Task, len(tasks))
	for i := range tasks {
		tasksMap[tasks[i].ID] = tasks[i]
	}
	if neededTask, ok := tasksMap[taskID]; ok {
		return neededTask, nil
	} else {
		return task.Task{}, ErrTaskNotFound
	}
}

func (s *Store) SetDone(taskID int, done bool) (task.Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	allTasks, err := s.load()
	if err != nil {
		return task.Task{}, err
	}
	for i := range allTasks {
		if allTasks[i].ID == taskID {
			allTasks[i].Done = done
			return allTasks[i], s.save(allTasks)
		}
	}
	return task.Task{}, ErrTaskNotFound
}

func (s *Store) Delete(taskID int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	allTasks, err := s.load()
	if err != nil {
		return err
	}
	idx := -1
	for i := range allTasks {
		if allTasks[i].ID == taskID {
			idx = i
			break
		}
	}
	if idx == -1 {
		return ErrTaskNotFound
	} else {
		newTasks := allTasks[:idx]
		newTasks = append(newTasks, allTasks[idx+1:]...)
		return s.save(newTasks)
	}
}

func (s *Store) Edit(t task.Task) (task.Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	allTasks, err := s.load()
	if err != nil {
		return task.Task{}, err
	}
	for i := range allTasks {
		if t.ID == allTasks[i].ID {
			allTasks[i] = t
			return t, s.save(allTasks)
		}
	}
	return task.Task{}, ErrTaskNotFound
}

func (s *Store) List() ([]task.Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.load()
}
