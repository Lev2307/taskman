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
	Edit(t task.Task) error
	Delete(taskID int) error
	SetDone(taskID int, done bool) error
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

func (c *Client) Edit(t task.Task) error {
	return fmt.Errorf("not implemented")
}

func (c *Client) Delete(taskID int) error {
	return fmt.Errorf("not implemented")
}

func (c *Client) SetDone(taskID int, done bool) error {
	return fmt.Errorf("not implemented")
}
