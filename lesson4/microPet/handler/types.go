package handler

import (
	"sync"

	"github.com/gorilla/mux"
)

type Task struct {
	Id    int    `json:"id"`
	Title string `json:"title"`
	Done  bool   `json:"done"`
}

type Error struct {
	Text string `json:"error"`
}

type Handler struct {
	Tasks  map[int]*Task
	Mu     *sync.Mutex
	Router *mux.Router
	LastID int
}
