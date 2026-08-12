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
	return mux
}
