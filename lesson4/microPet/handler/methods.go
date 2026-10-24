package handler

import (
	"bytes"
	"cmp"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.Router.ServeHTTP(w, r)
}

func (h *Handler) GetAllTasks(w http.ResponseWriter, r *http.Request) {
	var tasksIWantToSend = make([]Task, 0, 10)

	h.Mu.Lock()
	for _, task := range h.Tasks {
		tasksIWantToSend = append(tasksIWantToSend, *task)
	}
	h.Mu.Unlock()

	// Сортировка по возрастанию поля ID
	slices.SortFunc(tasksIWantToSend, func(a, b Task) int {
		return cmp.Compare(a.Id, b.Id)
	})

	var buffer bytes.Buffer

	err := json.NewEncoder(&buffer).Encode(tasksIWantToSend)
	if err != nil {
		responseError := Error{Text: "Server Internal Error"}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		err = json.NewEncoder(w).Encode(responseError)

		if err != nil {
			fmt.Println("Не получилось отправить ошибку:", err)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(buffer.Bytes())
	if err != nil {
		fmt.Println("Не получилось отправить ответ:", err)
	}
	return
}

func (h *Handler) GetCertainTask(w http.ResponseWriter, r *http.Request) {
	taskID, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil || taskID <= 0 {
		http.Error(w, "Bad field ID", http.StatusBadRequest)
		return
	}

	h.Mu.Lock()
	var taskWeNeedToReturn, ok = h.Tasks[taskID]

	if ok == false {
		h.Mu.Unlock()
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	h.Mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	err = json.NewEncoder(w).Encode(taskWeNeedToReturn)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

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
	if strings.TrimSpace(infoFromRequest.Title) == "" {
		http.Error(w, "Bad field Title", http.StatusBadRequest)
		return
	}

	h.Mu.Lock()

	h.LastID++
	// Создание новой задачи
	newTask := &Task{
		Id:    h.LastID,
		Title: infoFromRequest.Title,
		Done:  false,
	}
	h.Tasks[newTask.Id] = newTask

	h.Mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	err = json.NewEncoder(w).Encode(newTask)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

}

func (h *Handler) ChangeTask(w http.ResponseWriter, r *http.Request) {

	taskID, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil || taskID <= 0 {
		http.Error(w, "Bad field ID", http.StatusBadRequest)
		return
	}

	var infoToChangeStruct struct {
		Title *string `json:"title"`
		Done  *bool   `json:"done"`
	}

	bodyRequest := r.Body
	defer bodyRequest.Close()

	err = json.NewDecoder(bodyRequest).Decode(&infoToChangeStruct)
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	if infoToChangeStruct.Title == nil || infoToChangeStruct.Done == nil || strings.TrimSpace(*infoToChangeStruct.Title) == "" {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	h.Mu.Lock()
	var certainTaskToChange, ok = h.Tasks[taskID]
	if !ok {
		h.Mu.Unlock()
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	certainTaskToChange.Title = *infoToChangeStruct.Title
	certainTaskToChange.Done = *infoToChangeStruct.Done
	h.Mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	err = json.NewEncoder(w).Encode(certainTaskToChange)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusOK)
		return
	}

}

func (h *Handler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	taskID, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil || taskID <= 0 {
		http.Error(w, "Bad field ID", http.StatusBadRequest)
		return
	}

	h.Mu.Lock()

	var _, ok = h.Tasks[taskID]

	if !ok {
		h.Mu.Unlock()
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}
	delete(h.Tasks, taskID)
	h.Mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)

}
