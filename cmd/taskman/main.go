package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

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
		log.Println("listening on :8080")
		if err := http.ListenAndServe(":8080", srv.Routes()); err != nil {
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
