package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	task "github.com/Lev2307/taskman/internal/model"
	storage "github.com/Lev2307/taskman/internal/storage"
)

func (srv *Server) handleListTasks(w http.ResponseWriter, r *http.Request) {
	tasks, err := srv.store.List()
	if err != nil {
		log.Printf("list tasks: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	if tasks == nil {
		tasks = []task.Task{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}

func (srv *Server) handleDetailTask(w http.ResponseWriter, r *http.Request) {
	taskID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "id must be an integer", http.StatusBadRequest)
		return
	}
	t, err := srv.store.GetByID(taskID)
	if errors.Is(err, storage.ErrTaskNotFound) {
		http.Error(w, "task with given id not found", http.StatusNotFound)
		return
	} else if err != nil {
		log.Printf("detail task: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(t)
}

func (srv *Server) handleDeleteTask(w http.ResponseWriter, r *http.Request) {
	taskID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "id must be an integer", http.StatusBadRequest)
		return
	}
	deleteErr := srv.store.Delete(taskID)
	if errors.Is(deleteErr, storage.ErrTaskNotFound) {
		http.Error(w, "task with given id not found", http.StatusNotFound)
		return
	} else if deleteErr != nil {
		log.Printf("edit task: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode("task was deleted successfully")
}

func (srv *Server) handleEditTask(w http.ResponseWriter, r *http.Request) {
	taskUrlID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "id must be an integer", http.StatusBadRequest)
		return
	}
	var t task.Task
	if t.CreatedAt.IsZero() {
		t.CreatedAt = time.Now()
	}
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if taskUrlID != t.ID {
		http.Error(w, "discrepancies with JSON and URL task id", http.StatusBadRequest)
		return
	}
	editErr := srv.store.Edit(t)
	if editErr != nil {
		if errors.Is(editErr, storage.ErrTaskNotFound) {
			http.Error(w, "task with given id not found", http.StatusNotFound)
			return
		}
		log.Printf("edit task: %v", editErr)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(fmt.Sprintf("your task with id - %d was edited successfully", t.ID))
}

func (srv *Server) handleToggleDoneTask(w http.ResponseWriter, r *http.Request) {
	taskID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "id must be an integer", http.StatusBadRequest)
		return
	}
	var body struct {
		Done bool `json:"done"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if err := srv.store.SetDone(taskID, body.Done); err != nil {
		if errors.Is(err, storage.ErrTaskNotFound) {
			http.Error(w, "task with given id not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	s := "You`r task is now INcompleted!"
	if body.Done {
		s = "You`r task is now COMPLETED!"
	}
	json.NewEncoder(w).Encode(s)
}
