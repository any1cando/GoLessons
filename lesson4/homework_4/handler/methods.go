package handler

import (
	"cmp"
	"net/http"
	"slices"
)

func (handler *HandlerForHW) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handler.Router.ServeHTTP(w, r)
}

func (handler *HandlerForHW) ShowAllTasks() []TaskForHW {
	handler.Mu.Lock()

	tasks := make([]TaskForHW, 0, len(handler.Tasks))
	for _, task := range handler.Tasks {
		tasks = append(tasks, *task)
	}

	handler.Mu.Unlock()

	slices.SortFunc(tasks, func(a, b TaskForHW) int {
		return cmp.Compare(a.ID, b.ID)
	})

	return tasks
}

func (handler *HandlerForHW) CreateNewTask(title string, ownerID int64) *TaskForHW {

	handler.Mu.Lock()

	handler.LastID++
	newTask := &TaskForHW{
		ID:         handler.LastID,
		OwnerID:    ownerID,
		AssigneeID: 0,
		Title:      title,
		Status:     false,
	}
	handler.Tasks[newTask.ID] = newTask

	handler.Mu.Unlock()

	return newTask
}

func (handler *HandlerForHW) AssigneeTask(taskID int, assigneeID int64) bool {

	handler.Mu.Lock()
	task, ok := handler.Tasks[taskID]
	if !ok {
		handler.Mu.Unlock()
		return false
	}
	task.AssigneeID = assigneeID
	handler.Mu.Unlock()

	return true
}
