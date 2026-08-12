package api

import (
	"encoding/json"
	"log"
	"net/http"

	task "github.com/Lev2307/taskman/internal/model"
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
