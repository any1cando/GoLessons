package handler

import (
	"sync"

	"github.com/gorilla/mux"
)

type TaskForHW struct {
	ID         int    `json:"id"`
	OwnerID    int64  `json:"owner_id"`
	AssigneeID int64  `json:"assignee_id"`
	Title      string `json:"title"`
	Status     bool   `json:"status"`
}

type ErrorHW struct {
	Title string `json:"error"`
}

type HandlerForHW struct {
	Tasks  map[int]*TaskForHW
	Mu     *sync.Mutex
	Router *mux.Router
	LastID int // Индекс для задач, чтобы удалять их без проблем без смещений списка
}
