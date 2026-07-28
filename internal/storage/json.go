package storage

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	task "github.com/Lev2307/taskman/internal/model"
)

func LoadJson(path string) ([]task.Task, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []task.Task{}, nil
		}
		return []task.Task{}, fmt.Errorf("open file: %w", err)
	}
	defer f.Close()

	var tasks []task.Task
	if err := json.NewDecoder(f).Decode(&tasks); err != nil {
		if err == io.EOF {
			return []task.Task{}, nil
		}
		return []task.Task{}, fmt.Errorf("decode: %w", err)
	}
	return tasks, nil
}

func SaveJson(path string, tasks []task.Task) error {
	f, err := os.Create(path)
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

func AddTask(path string, task task.Task) error {
	tasks, err := LoadJson(path)
	if err != nil {
		return err
	}
	tasks = append(tasks, task)
	return SaveJson(path, tasks)
}

func GetLastID(path string) (int, error) {
	data, err := LoadJson(path)
	if err != nil {
		return 0, fmt.Errorf("Err with file: %w", err)
	}
	last_data_element := data[len(data)-1]
	return last_data_element.ID, nil
}
