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

func EncodeMessageAndCheckError(w http.ResponseWriter, statusCode int, infoToEncode any) {
	var buffer bytes.Buffer

	err := json.NewEncoder(&buffer).Encode(infoToEncode)
	if err != nil {
		responseError := Error{Text: "Internal Server Error"}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)

		err = json.NewEncoder(w).Encode(responseError)
		if err != nil {
			fmt.Println("Ошибку не удалось отправить")
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	_, err = w.Write(buffer.Bytes())
	if err != nil {
		fmt.Println("Не получилось отправить ответ:", err)
	}
}

func SecondDecoderCheck(w http.ResponseWriter, bodyRequest io.ReadCloser, structToDecode any) bool {
	decoder := json.NewDecoder(bodyRequest)

	err := decoder.Decode(structToDecode)
	if err != nil {
		EncodeMessageAndCheckError(w, http.StatusBadRequest, Error{Text: "Bad Request"})
		return false
	}

	var extraCheckDecode any
	if err = decoder.Decode(&extraCheckDecode); err != io.EOF {
		EncodeMessageAndCheckError(w, http.StatusBadRequest, Error{Text: "Bad Request"})
		return false
	}
	return true
}

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

	EncodeMessageAndCheckError(w, http.StatusOK, tasksIWantToSend)
}

func (h *Handler) GetCertainTask(w http.ResponseWriter, r *http.Request) {
	taskID, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil || taskID <= 0 {
		EncodeMessageAndCheckError(w, http.StatusBadRequest, Error{Text: "Bad Request"})
		return
	}

	h.Mu.Lock()
	var taskWeNeedToReturn, ok = h.Tasks[taskID]

	if ok == false {
		h.Mu.Unlock()
		EncodeMessageAndCheckError(w, http.StatusNotFound, Error{Text: "Not Found"})
		return
	}

	copyOfTask := *taskWeNeedToReturn
	h.Mu.Unlock()

	EncodeMessageAndCheckError(w, http.StatusOK, copyOfTask)

}

func (h *Handler) CreateTask(w http.ResponseWriter, r *http.Request) {
	// Создаем структуру для распарса только 2/3 полей. Третье поле мы сами создадим (это id)
	var infoFromRequest struct {
		Title string `json:"title"`
	}
	var bodyRequest io.ReadCloser = r.Body
	defer bodyRequest.Close()

	if SecondDecoderCheck(w, bodyRequest, &infoFromRequest) == false {
		return
	}

	// Проверяем валидность полей
	if strings.TrimSpace(infoFromRequest.Title) == "" {
		EncodeMessageAndCheckError(w, http.StatusBadRequest, Error{Text: "Bad field Title"})
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

	copyTask := *newTask
	h.Mu.Unlock()

	EncodeMessageAndCheckError(w, http.StatusCreated, copyTask)

}

func (h *Handler) ChangeTask(w http.ResponseWriter, r *http.Request) {

	taskID, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil || taskID <= 0 {
		EncodeMessageAndCheckError(w, http.StatusBadRequest, Error{Text: "Bad field ID"})
		return
	}

	var infoToChangeStruct struct {
		Title *string `json:"title"`
		Done  *bool   `json:"done"`
	}

	bodyRequest := r.Body
	defer bodyRequest.Close()

	if SecondDecoderCheck(w, bodyRequest, &infoToChangeStruct) == false {
		return
	}

	if infoToChangeStruct.Title == nil || infoToChangeStruct.Done == nil || strings.TrimSpace(*infoToChangeStruct.Title) == "" {
		EncodeMessageAndCheckError(w, http.StatusBadRequest, Error{Text: "Bad Request"})
		return
	}

	h.Mu.Lock()
	var certainTaskToChange, ok = h.Tasks[taskID]
	if !ok {
		h.Mu.Unlock()
		EncodeMessageAndCheckError(w, http.StatusNotFound, Error{Text: "Not Found"})
		return
	}

	certainTaskToChange.Title = *infoToChangeStruct.Title
	certainTaskToChange.Done = *infoToChangeStruct.Done

	copyOfCertainTask := *certainTaskToChange
	h.Mu.Unlock()

	EncodeMessageAndCheckError(w, http.StatusOK, copyOfCertainTask)

}

func (h *Handler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	taskID, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil || taskID <= 0 {
		EncodeMessageAndCheckError(w, http.StatusBadRequest, Error{Text: "Bad field ID"})
		return
	}

	h.Mu.Lock()

	var _, ok = h.Tasks[taskID]

	if !ok {
		h.Mu.Unlock()
		EncodeMessageAndCheckError(w, http.StatusNotFound, Error{Text: "Not Found"})
		return
	}
	delete(h.Tasks, taskID)
	h.Mu.Unlock()

	// Не вызываем функцию, которая заворачивает сообщение в JSON, потому что и заворачивать нечего
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)

}
