package handler

import (
	"encoding/json"
	"io"
	"net/http"
)

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.Router.ServeHTTP(w, r)
}

func (h *Handler) GetAllTasks(w http.ResponseWriter, r *http.Request) {
	var tasksIWantToSend []*Task

	h.Mu.Lock()
	for _, task := range h.Tasks {
		tasksIWantToSend = append(tasksIWantToSend, task)
	}
	h.Mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	err := json.NewEncoder(w).Encode(tasksIWantToSend)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

}

func (h *Handler) GetCertainTask(w http.ResponseWriter, r *http.Request) {

}

func (h *Handler) CreateTask(w http.ResponseWriter, r *http.Request) {
	// Создаем структуру для распарса только 2/3 полей. Третье поле мы сами создадим (это id)
	var infoFromRequest struct {
		Title string `json:"title"`
	}
	var bodyRequest io.ReadCloser = r.Body
	defer bodyRequest.Close()

	err := json.NewDecoder(bodyRequest).Decode(&infoFromRequest)
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// Проверяем валидность полей
	if infoFromRequest.Title == "" {
		http.Error(w, "Bad field Title", http.StatusBadRequest)
		return
	}

	h.Mu.Lock()

	// Создание новой задачи
	newTask := &Task{
		Id:    len(h.Tasks) + 1,
		Title: infoFromRequest.Title,
		Done:  false,
	}
	h.Tasks[newTask.Id] = newTask

	h.Mu.Unlock()
	err = json.NewEncoder(w).Encode(newTask)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

}

func (h *Handler) ChangeTask(w http.ResponseWriter, r *http.Request) {

}

func (h *Handler) DeleteTask(w http.ResponseWriter, r *http.Request) {

}

// Добавить и другие методы
