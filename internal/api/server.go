package api

import (
	"net/http"

	storage "github.com/Lev2307/taskman/internal/storage"
)

type Server struct {
	store *storage.Store
}

func NewServer(s *storage.Store) *Server {
	return &Server{store: s}
}

func (srv *Server) Routes() http.Handler {
	mux := http.NewServeMux() // mux автоматически отсеивает неиспользуемые методы для url: +rep
	mux.HandleFunc("GET /tasks", srv.handleListTasks)
	mux.HandleFunc("POST /tasks", srv.handleAddTask)
	mux.HandleFunc("GET /tasks/{id}", srv.handleDetailTask)
	mux.HandleFunc("DELETE /tasks/{id}", srv.handleDeleteTask)
	mux.HandleFunc("PUT /tasks/{id}", srv.handleEditTask)
	mux.HandleFunc("PUT /tasks/{id}/done", srv.handleToggleDoneTask)
	return mux
}
