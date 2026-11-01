package handler

import (
	"cmp"
	"net/http"
	"slices"
)

func sortedSliceWithTasks(tasks []TaskForHW) []TaskForHW {
	slices.SortFunc(tasks, func(a, b TaskForHW) int {
		return cmp.Compare(a.ID, b.ID)
	})
	return tasks
}

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

	return sortedSliceWithTasks(tasks)
}

func (handler *HandlerForHW) CreateNewTask(title string, ownerID int64) TaskForHW {

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

	return *newTask
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

func (handler *HandlerForHW) UnassigneeTask(taskID int) bool {

	handler.Mu.Lock()
	task, ok := handler.Tasks[taskID]
	if !ok {
		handler.Mu.Unlock()
		return false
	}
	task.AssigneeID = 0
	handler.Mu.Unlock()

	return true
}

func (handler *HandlerForHW) ResolveTask(taskID int) bool {

	handler.Mu.Lock()
	task, ok := handler.Tasks[taskID]
	if !ok {
		handler.Mu.Unlock()
		return false
	}
	task.Status = true
	delete(handler.Tasks, task.ID)
	handler.Mu.Unlock()

	return true
}

func (handler *HandlerForHW) ShowUserHisAssigneeTasks(assigneeID int64) []TaskForHW {

	handler.Mu.Lock()
	tasks := make([]TaskForHW, 0, len(handler.Tasks))
	for _, task := range handler.Tasks {
		if task.AssigneeID == assigneeID {
			tasks = append(tasks, *task)
		}
	}
	handler.Mu.Unlock()
	return sortedSliceWithTasks(tasks)

}

func (handler *HandlerForHW) ShowUserHisOwnTasks(ownerID int64) []TaskForHW {

	handler.Mu.Lock()
	tasks := make([]TaskForHW, 0, len(handler.Tasks))
	for _, task := range handler.Tasks {
		if task.OwnerID == ownerID {
			tasks = append(tasks, *task)
		}
	}
	handler.Mu.Unlock()
	return sortedSliceWithTasks(tasks)

}

func (handler *HandlerForHW) MakeTaskDone(taskID int) bool {

	handler.Mu.Lock()
	task, ok := handler.Tasks[taskID]
	if !ok {
		handler.Mu.Unlock()
		return false
	}
	task.Status = true
	handler.Mu.Unlock()

	return true
}

func (handler *HandlerForHW) MakeTaskUndone(taskID int) bool {

	handler.Mu.Lock()
	task, ok := handler.Tasks[taskID]
	if !ok {
		handler.Mu.Unlock()
		return false
	}
	task.Status = false
	handler.Mu.Unlock()

	return true
}

// TODO: Доделать команды done_ID и undone_ID
