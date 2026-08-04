package storage

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	task "github.com/Lev2307/taskman/internal/model"
)

func LoadTasksJson(path string) ([]task.Task, error) {
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

func SaveTasksJson(path string, tasks []task.Task) error {
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
	tasks, err := LoadTasksJson(path)
	if err != nil {
		return err
	}
	tasks = append(tasks, task)
	return SaveTasksJson(path, tasks)
}

func GetTaskByID(path string, taskID int) (task.Task, error) {
	tasks, err := LoadTasksJson(path)
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
		return task.Task{}, fmt.Errorf("task with given id was not found")
	}
}

func GetLastID(path string) (int, error) {
	data, err := LoadTasksJson(path)
	if err != nil {
		return 0, fmt.Errorf("err with file: %w", err)
	}
	if len(data) == 0 {
		return 0, nil
	}
	last_data_element := data[len(data)-1]
	return last_data_element.ID, nil
}

func ToggleDone(path string, taskID int) error {
	allTasks, err := LoadTasksJson(path)
	if err != nil {
		return err
	}
	for i := range allTasks {
		if allTasks[i].ID == taskID {
			allTasks[i].Done = !allTasks[i].Done
		}
	}
	return SaveTasksJson(path, allTasks)
}

func DeleteTask(path string, taskID int) error {
	allTasks, err := LoadTasksJson(path)
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
		return fmt.Errorf("task with given id was not found")
	} else {
		newTasks := allTasks[:idx]
		newTasks = append(newTasks, allTasks[idx+1:]...)
		return SaveTasksJson(path, newTasks)
	}
}
