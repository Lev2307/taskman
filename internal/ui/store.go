package ui

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	task "github.com/Lev2307/taskman/internal/model"
)

type TaskStore interface {
	List() ([]task.Task, error)
	GetByID(taskID int) (task.Task, error)
	Add(t task.Task) (task.Task, error)
	Edit(t task.Task) (task.Task, error)
	Delete(taskID int) error
	SetDone(taskID int, done bool) (task.Task, error)
}

type Client struct {
	baseURL string
	http    *http.Client
}

var _ TaskStore = (*Client)(nil)

func NewClient(baseURL string) *Client {
	return &Client{baseURL: baseURL, http: &http.Client{Timeout: 5 * time.Second}}
}

func (c *Client) errFromResponse(r *http.Response) error {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("read error body: %w", err)
	}
	return fmt.Errorf("%s: %s", r.Status, strings.TrimSpace(string(data)))
}

func (c *Client) List() ([]task.Task, error) {
	url := c.baseURL + "/tasks"
	response, err := c.http.Get(url)
	if err != nil {
		return []task.Task{}, fmt.Errorf("api list task error: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return []task.Task{}, c.errFromResponse(response)
	}
	var t []task.Task
	if err := json.NewDecoder(response.Body).Decode(&t); err != nil {
		return []task.Task{}, fmt.Errorf("api list task decoding json error: %w", err)
	}
	return t, nil
}

func (c *Client) GetByID(taskID int) (task.Task, error) {
	url := fmt.Sprintf("%s/tasks/%d", c.baseURL, taskID)
	response, err := c.http.Get(url)
	if err != nil {
		return task.Task{}, fmt.Errorf("api detail task error: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return task.Task{}, c.errFromResponse(response)
	}

	var t task.Task
	if err := json.NewDecoder(response.Body).Decode(&t); err != nil {
		return task.Task{}, fmt.Errorf("api list task decoding json error: %w", err)
	}
	return t, nil
}

func (c *Client) Add(t task.Task) (task.Task, error) {
	url := c.baseURL + "/tasks"
	jsonData, err := json.Marshal(t)
	if err != nil {
		return task.Task{}, fmt.Errorf("error marshaling JSON: %w", err)
	}
	response, err := c.http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return task.Task{}, fmt.Errorf("api add task error: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusCreated {
		return task.Task{}, c.errFromResponse(response)
	}

	var createdTask task.Task
	if err := json.NewDecoder(response.Body).Decode(&createdTask); err != nil {
		return task.Task{}, fmt.Errorf("api add task decoding json error: %w", err)
	}
	return createdTask, nil
}

func (c *Client) Edit(t task.Task) (task.Task, error) {
	url := fmt.Sprintf("%s/tasks/%d", c.baseURL, t.ID)
	jsonData, err := json.Marshal(t)
	if err != nil {
		return task.Task{}, fmt.Errorf("error marshaling JSON: %w", err)
	}
	request, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return task.Task{}, fmt.Errorf("error invalid request: %w", err)
	}
	response, err := c.http.Do(request)
	if err != nil {
		return task.Task{}, fmt.Errorf("error invalid response: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return task.Task{}, c.errFromResponse(response)
	}

	var editedTask task.Task
	if err := json.NewDecoder(response.Body).Decode(&editedTask); err != nil {
		return task.Task{}, fmt.Errorf("api edit task decoding json error: %w", err)
	}
	return editedTask, nil
}

func (c *Client) Delete(taskID int) error {
	url := fmt.Sprintf("%s/tasks/%d", c.baseURL, taskID)
	request, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return fmt.Errorf("error invalid request: %w", err)
	}
	response, err := c.http.Do(request)
	if err != nil {
		return fmt.Errorf("error invalid response: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusNoContent {
		return c.errFromResponse(response)
	}
	return nil
}

func (c *Client) SetDone(taskID int, done bool) (task.Task, error) {
	url := fmt.Sprintf("%s/tasks/%d/done", c.baseURL, taskID)
	payload, err := json.Marshal(map[string]bool{"done": done})
	if err != nil {
		return task.Task{}, fmt.Errorf("error marshaling JSON: %w", err)
	}
	request, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(payload))
	if err != nil {
		return task.Task{}, fmt.Errorf("error invalid request: %w", err)
	}
	response, err := c.http.Do(request)
	if err != nil {
		return task.Task{}, fmt.Errorf("error invalid response: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return task.Task{}, c.errFromResponse(response)
	}

	var toggledTask task.Task
	if err := json.NewDecoder(response.Body).Decode(&toggledTask); err != nil {
		return task.Task{}, fmt.Errorf("api toggling task decoding json error: %w", err)
	}
	return toggledTask, nil
}
