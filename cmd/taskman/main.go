package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	api "github.com/Lev2307/taskman/internal/api"
	storage "github.com/Lev2307/taskman/internal/storage"
	ui "github.com/Lev2307/taskman/internal/ui"
	tea "github.com/charmbracelet/bubbletea"
)

const PATH string = "tasks.json"

func main() {
	store := storage.NewStore(PATH)
	serve := flag.Bool("serve", false, "запустить HTTP-сервер вместо TUI")
	flag.Parse()
	if *serve {
		srv := api.NewServer(store)
		httpServ := &http.Server{Addr: ":8080", Handler: srv.Routes(), ReadTimeout: 5 * time.Second, WriteTimeout: 10 * time.Second}
		log.Println("listening on :8080")
		if err := httpServ.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
		return
	}
	app := ui.New(store)
	if _, err := tea.NewProgram(app).Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
