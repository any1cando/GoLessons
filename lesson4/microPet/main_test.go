package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync"
	"testing"

	"firstGoProject/lesson4/microPet/handler"

	"github.com/gorilla/mux"
)

// Для каждого теста создаём отдельное пустое хранилище.
func newTestAPI() http.Handler {
	router := mux.NewRouter().StrictSlash(true)

	api := &handler.Handler{
		Tasks:  make(map[int]*handler.Task),
		Mu:     &sync.Mutex{},
		Router: router,
		LastID: 0,
	}

	router.HandleFunc("/tasks/", api.GetAllTasks).Methods("GET")
	router.HandleFunc("/tasks/{id}", api.GetCertainTask).Methods("GET")
	router.HandleFunc("/tasks/", api.CreateTask).Methods("POST")
	router.HandleFunc("/tasks/{id}", api.ChangeTask).Methods("PUT")
	router.HandleFunc("/tasks/{id}", api.DeleteTask).Methods("DELETE")

	router.NotFoundHandler = http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			handler.EncodeMessageAndCheckError(
				w,
				http.StatusNotFound,
				handler.Error{Text: "Not Found"},
			)
		},
	)

	router.MethodNotAllowedHandler = http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			handler.EncodeMessageAndCheckError(
				w,
				http.StatusMethodNotAllowed,
				handler.Error{Text: "Method Not Allowed"},
			)
		},
	)

	return api
}

func request(
	api http.Handler,
	method string,
	path string,
	body string,
) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))

	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}

	rec := httptest.NewRecorder()
	api.ServeHTTP(rec, req)
	return rec
}

func requireStatus(
	t *testing.T,
	rec *httptest.ResponseRecorder,
	want int,
) {
	t.Helper()

	if rec.Code != want {
		t.Fatalf(
			"статус: получили %d, хотели %d; тело: %s",
			rec.Code,
			want,
			rec.Body.String(),
		)
	}
}

func requireJSON(t *testing.T, rec *httptest.ResponseRecorder) {
	t.Helper()

	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("Content-Type: получили %q", got)
	}

	if !json.Valid(rec.Body.Bytes()) {
		t.Fatalf("ответ не является корректным JSON: %q", rec.Body.String())
	}
}

func requireError(
	t *testing.T,
	rec *httptest.ResponseRecorder,
	status int,
) {
	t.Helper()

	requireStatus(t, rec, status)
	requireJSON(t, rec)

	var result map[string]json.RawMessage
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("ожидали JSON-объект ошибки: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("ожидали только поле error, получили: %s", rec.Body.String())
	}

	var message string
	if err := json.Unmarshal(result["error"], &message); err != nil {
		t.Fatalf("поле error должно быть строкой: %v", err)
	}

	if strings.TrimSpace(message) == "" {
		t.Fatal("пустое описание ошибки")
	}
}

func decodeTask(t *testing.T, rec *httptest.ResponseRecorder) handler.Task {
	t.Helper()
	requireJSON(t, rec)

	// Проверяем не только значения, но и наличие нужных JSON-полей.
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(rec.Body.Bytes(), &fields); err != nil {
		t.Fatalf("ответ должен быть объектом задачи: %v", err)
	}

	if len(fields) != 3 {
		t.Fatalf("ожидали поля id, title, done: %s", rec.Body.String())
	}

	for _, name := range []string{"id", "title", "done"} {
		value, ok := fields[name]
		if !ok || string(value) == "null" {
			t.Fatalf("отсутствует значение поля %q: %s", name, rec.Body.String())
		}
	}

	var task handler.Task
	if err := json.Unmarshal(rec.Body.Bytes(), &task); err != nil {
		t.Fatalf("не удалось прочитать задачу: %v", err)
	}

	if task.Id <= 0 {
		t.Fatalf("ID должен быть положительным: %d", task.Id)
	}

	return task
}

func createTask(t *testing.T, api http.Handler, title string) handler.Task {
	t.Helper()

	body, err := json.Marshal(map[string]string{"title": title})
	if err != nil {
		t.Fatal(err)
	}

	rec := request(api, http.MethodPost, "/tasks/", string(body))
	requireStatus(t, rec, http.StatusCreated)
	task := decodeTask(t, rec)

	if task.Title != title || task.Done {
		t.Fatalf("неверная созданная задача: %+v", task)
	}

	return task
}

func getTask(t *testing.T, api http.Handler, id int) handler.Task {
	t.Helper()

	rec := request(api, http.MethodGet, fmt.Sprintf("/tasks/%d", id), "")
	requireStatus(t, rec, http.StatusOK)
	return decodeTask(t, rec)
}

func listTasks(t *testing.T, api http.Handler) []handler.Task {
	t.Helper()

	rec := request(api, http.MethodGet, "/tasks/", "")
	requireStatus(t, rec, http.StatusOK)
	requireJSON(t, rec)

	var tasks []handler.Task
	if err := json.Unmarshal(rec.Body.Bytes(), &tasks); err != nil {
		t.Fatalf("не удалось прочитать массив задач: %v", err)
	}

	if tasks == nil {
		t.Fatal("ожидали массив, получили null")
	}

	for i := 1; i < len(tasks); i++ {
		if tasks[i-1].Id >= tasks[i].Id {
			t.Fatalf("список не отсортирован или ID повторяются: %+v", tasks)
		}
	}

	return tasks
}

func TestEmptyList(t *testing.T) {
	api := newTestAPI()

	rec := request(api, http.MethodGet, "/tasks/", "")
	requireStatus(t, rec, http.StatusOK)
	requireJSON(t, rec)

	if got := strings.TrimSpace(rec.Body.String()); got != "[]" {
		t.Fatalf("ожидали [], получили %q", got)
	}
}

func TestTaskLifecycle(t *testing.T) {
	api := newTestAPI()

	created := createTask(t, api, "Убрать дома")

	if got := getTask(t, api, created.Id); got != created {
		t.Fatalf("полученная задача отличается: %+v", got)
	}

	// Отмечаем выполненной.
	path := fmt.Sprintf("/tasks/%d", created.Id)
	rec := request(api, http.MethodPut, path,
		`{"title":"Убрать дома и на балконе","done":true}`)
	requireStatus(t, rec, http.StatusOK)

	updated := decodeTask(t, rec)
	want := handler.Task{
		Id:    created.Id,
		Title: "Убрать дома и на балконе",
		Done:  true,
	}

	if updated != want {
		t.Fatalf("получили %+v, хотели %+v", updated, want)
	}

	if got := getTask(t, api, created.Id); got != want {
		t.Fatalf("изменение не сохранилось: %+v", got)
	}

	// Выполненная задача должна оставаться в списке.
	tasks := listTasks(t, api)
	if len(tasks) != 1 || tasks[0] != want {
		t.Fatalf("неверный список после выполнения: %+v", tasks)
	}

	// Явное false должно снять отметку.
	rec = request(api, http.MethodPut, path,
		`{"title":"Убрать дома","done":false}`)
	requireStatus(t, rec, http.StatusOK)

	updated = decodeTask(t, rec)
	if updated.Id != created.Id || updated.Done || updated.Title != "Убрать дома" {
		t.Fatalf("неверное обновление с false: %+v", updated)
	}

	if got := getTask(t, api, created.Id); got != updated {
		t.Fatalf("обновление с false не сохранилось: %+v", got)
	}

	// Удаление: 204, пустое тело.
	rec = request(api, http.MethodDelete, path, "")
	requireStatus(t, rec, http.StatusNoContent)

	if rec.Body.Len() != 0 {
		t.Fatalf("ответ 204 должен быть пустым: %q", rec.Body.String())
	}

	requireError(t, request(api, http.MethodGet, path, ""), http.StatusNotFound)
	requireError(t, request(api, http.MethodDelete, path, ""), http.StatusNotFound)

	if tasks := listTasks(t, api); len(tasks) != 0 {
		t.Fatalf("после удаления остались задачи: %+v", tasks)
	}
}

func TestListOrderAndUniqueIDs(t *testing.T) {
	api := newTestAPI()
	var created []handler.Task

	for i := 0; i < 30; i++ {
		created = append(created, createTask(t, api, fmt.Sprintf("Задача %d", i)))
	}

	// Несколько чтений, чтобы проверять порядок независимо от обхода map.
	for i := 0; i < 10; i++ {
		got := listTasks(t, api)
		if !reflect.DeepEqual(got, created) {
			t.Fatalf("список отличается от созданных задач:\n%+v", got)
		}
	}
}

func TestIDsAreNotReused(t *testing.T) {
	for _, deletedIndex := range []int{0, 1, 2} {
		t.Run(fmt.Sprintf("delete_index_%d", deletedIndex), func(t *testing.T) {
			api := newTestAPI()
			created := []handler.Task{
				createTask(t, api, "Первая"),
				createTask(t, api, "Вторая"),
				createTask(t, api, "Третья"),
			}

			deleted := created[deletedIndex]
			rec := request(api, http.MethodDelete,
				fmt.Sprintf("/tasks/%d", deleted.Id), "")
			requireStatus(t, rec, http.StatusNoContent)

			newTask := createTask(t, api, "Новая")
			if newTask.Id <= created[2].Id {
				t.Fatalf("ID использован повторно: %+v", newTask)
			}

			for i, original := range created {
				if i == deletedIndex {
					continue
				}
				if got := getTask(t, api, original.Id); got != original {
					t.Fatalf("существующая задача изменилась: %+v", got)
				}
			}

			if got := listTasks(t, api); len(got) != 3 {
				t.Fatalf("ожидали 3 задачи, получили %d", len(got))
			}

			// Даже после удаления всех задач счётчик не сбрасывается.
			for _, task := range listTasks(t, api) {
				rec := request(api, http.MethodDelete,
					fmt.Sprintf("/tasks/%d", task.Id), "")
				requireStatus(t, rec, http.StatusNoContent)
			}

			next := createTask(t, api, "После удаления всех")
			if next.Id <= newTask.Id {
				t.Fatalf("счётчик ID сбросился: %+v", next)
			}
		})
	}
}

func TestCreateInvalidBody(t *testing.T) {
	cases := []struct {
		name string
		body string
	}{
		{"empty_body", ""},
		{"whitespace_body", " \n\t "},
		{"broken_json", `{"title":`},
		{"missing_title", `{}`},
		{"empty_title", `{"title":""}`},
		{"spaces_title", `{"title":"   "}`},
		{"unicode_spaces", `{"title":"\u00a0\t\n"}`},
		{"null_title", `{"title":null}`},
		{"number_title", `{"title":123}`},
		{"boolean_title", `{"title":true}`},
		{"array_title", `{"title":[]}`},
		{"object_title", `{"title":{}}`},
		{"null_body", `null`},
		{"array_body", `[]`},
		{"string_body", `"hello"`},
		{"number_body", `42`},
		{"trailing_comma", `{"title":"Task",}`},
		{"trailing_garbage", `{"title":"Task"} garbage`},
		{"second_object", `{"title":"Task"} {"title":"Other"}`},
		{"second_null", `{"title":"Task"} null`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			api := newTestAPI()
			rec := request(api, http.MethodPost, "/tasks/", tc.body)
			requireError(t, rec, http.StatusBadRequest)

			if got := listTasks(t, api); len(got) != 0 {
				t.Fatalf("невалидный запрос создал задачу: %+v", got)
			}
		})
	}
}

func TestUpdateInvalidBodyDoesNotChangeTask(t *testing.T) {
	cases := []struct {
		name string
		body string
	}{
		{"empty_body", ""},
		{"broken_json", `{"title":`},
		{"missing_both", `{}`},
		{"missing_done", `{"title":"Новое название"}`},
		{"missing_title", `{"done":false}`},
		{"null_title", `{"title":null,"done":false}`},
		{"null_done", `{"title":"Новое","done":null}`},
		{"empty_title", `{"title":"","done":false}`},
		{"spaces_title", `{"title":" \t\n","done":false}`},
		{"number_title", `{"title":123,"done":false}`},
		{"string_done", `{"title":"Новое","done":"false"}`},
		{"number_done", `{"title":"Новое","done":0}`},
		{"object_done", `{"title":"Новое","done":{}}`},
		{"array_done", `{"title":"Новое","done":[]}`},
		{"null_body", `null`},
		{"array_body", `[]`},
		{"trailing_garbage", `{"title":"Новое","done":false} garbage`},
		{"second_object", `{"title":"Новое","done":false} {}`},
		{"second_null", `{"title":"Новое","done":false} null`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			api := newTestAPI()
			created := createTask(t, api, "Исходная")
			path := fmt.Sprintf("/tasks/%d", created.Id)

			// Начинаем с done=true, чтобы заметить случайный сброс на false.
			rec := request(api, http.MethodPut, path,
				`{"title":"Исходная","done":true}`)
			requireStatus(t, rec, http.StatusOK)
			original := decodeTask(t, rec)

			rec = request(api, http.MethodPut, path, tc.body)
			requireError(t, rec, http.StatusBadRequest)

			if got := getTask(t, api, created.Id); got != original {
				t.Fatalf("невалидный запрос изменил задачу: %+v", got)
			}
		})
	}
}

func TestValidJSONWithTrailingWhitespace(t *testing.T) {
	api := newTestAPI()

	rec := request(api, http.MethodPost, "/tasks/",
		"{\"title\":\"Задача\"}\n \t")
	requireStatus(t, rec, http.StatusCreated)
	created := decodeTask(t, rec)

	rec = request(api, http.MethodPut,
		fmt.Sprintf("/tasks/%d", created.Id),
		"{\"title\":\"Новое название\",\"done\":true}\n \t")
	requireStatus(t, rec, http.StatusOK)

	if got := decodeTask(t, rec); got.Title != "Новое название" || !got.Done {
		t.Fatalf("неверное обновление: %+v", got)
	}
}

func TestSpecialCharactersInTitle(t *testing.T) {
	api := newTestAPI()

	for _, title := range []string{
		"Купить молоко 🥛",
		`Задача с "кавычками" и \ слешем`,
		"Первая строка\nВторая строка",
		"<script>alert(1)</script>",
	} {
		created := createTask(t, api, title)
		if got := getTask(t, api, created.Id); got.Title != title {
			t.Fatalf("название исказилось: получили %q, хотели %q", got.Title, title)
		}
	}
}

func TestCreateServerControlsIDAndDone(t *testing.T) {
	api := newTestAPI()

	rec := request(api, http.MethodPost, "/tasks/",
		`{"title":"Первая","id":999,"done":true}`)
	requireStatus(t, rec, http.StatusCreated)
	first := decodeTask(t, rec)

	if first.Done {
		t.Fatal("новая задача должна иметь done=false")
	}

	second := createTask(t, api, "Вторая")

	if first.Id != 1 || second.Id != 2 {
		t.Fatalf("ID должны назначаться сервером: %d, %d", first.Id, second.Id)
	}
}

func TestInvalidIDs(t *testing.T) {
	ids := []string{
		"0",
		"-1",
		"abc",
		"1.5",
		"999999999999999999999999999999999999",
	}

	for _, method := range []string{
		http.MethodGet,
		http.MethodPut,
		http.MethodDelete,
	} {
		for _, id := range ids {
			t.Run(method+"_"+id, func(t *testing.T) {
				api := newTestAPI()
				original := createTask(t, api, "Сохранить")

				body := ""
				if method == http.MethodPut {
					body = `{"title":"Изменение","done":true}`
				}

				rec := request(api, method, "/tasks/"+id, body)
				requireError(t, rec, http.StatusBadRequest)

				if got := getTask(t, api, original.Id); got != original {
					t.Fatalf("запрос с неверным ID изменил задачу: %+v", got)
				}
			})
		}
	}
}

func TestMissingTask(t *testing.T) {
	for _, method := range []string{
		http.MethodGet,
		http.MethodPut,
		http.MethodDelete,
	} {
		t.Run(method, func(t *testing.T) {
			api := newTestAPI()

			body := ""
			if method == http.MethodPut {
				body = `{"title":"Не существует","done":true}`
			}

			rec := request(api, method, "/tasks/999", body)
			requireError(t, rec, http.StatusNotFound)

			// Проверяем, что запрос не оставил мьютекс заблокированным.
			createTask(t, api, "После ошибки")
		})
	}
}

func TestUnknownRoutes(t *testing.T) {
	api := newTestAPI()

	for _, path := range []string{
		"/",
		"/unknown/",
		"/task/",
		"/tasks/1/extra",
	} {
		t.Run(path, func(t *testing.T) {
			requireError(t,
				request(api, http.MethodGet, path, ""),
				http.StatusNotFound,
			)
		})
	}
}

func TestUnsupportedMethods(t *testing.T) {
	api := newTestAPI()
	created := createTask(t, api, "Сохранить")

	cases := []struct {
		method string
		path   string
	}{
		{http.MethodPut, "/tasks/"},
		{http.MethodDelete, "/tasks/"},
		{http.MethodPatch, "/tasks/"},
		{http.MethodOptions, "/tasks/"},
		{http.MethodPost, fmt.Sprintf("/tasks/%d", created.Id)},
		{http.MethodPatch, fmt.Sprintf("/tasks/%d", created.Id)},
		{http.MethodOptions, fmt.Sprintf("/tasks/%d", created.Id)},
	}

	for _, tc := range cases {
		t.Run(tc.method+"_"+tc.path, func(t *testing.T) {
			rec := request(api, tc.method, tc.path, "")
			requireError(t, rec, http.StatusMethodNotAllowed)
		})
	}

	if got := getTask(t, api, created.Id); got != created {
		t.Fatalf("неподдерживаемый метод изменил задачу: %+v", got)
	}
}

func TestSeparateInstancesHaveSeparateStorage(t *testing.T) {
	firstAPI := newTestAPI()
	createTask(t, firstAPI, "Только в первом")

	secondAPI := newTestAPI()
	if got := listTasks(t, secondAPI); len(got) != 0 {
		t.Fatalf("новое хранилище не пустое: %+v", got)
	}
}

func TestEncodingFailureReturnsJSON500(t *testing.T) {
	rec := httptest.NewRecorder()

	// Каналы невозможно представить в JSON.
	handler.EncodeMessageAndCheckError(
		rec,
		http.StatusOK,
		make(chan int),
	)

	requireError(t, rec, http.StatusInternalServerError)
}

// Здесь каждый запрос создаёт свою задачу.
// Проверяем уникальность ID и отсутствие потерь.
func TestConcurrentCreate(t *testing.T) {
	api := newTestAPI()
	const count = 100

	results := make(chan *httptest.ResponseRecorder, count)
	var wg sync.WaitGroup

	for i := 0; i < count; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			body := fmt.Sprintf(`{"title":"Task %d"}`, i)
			results <- request(api, http.MethodPost, "/tasks/", body)
		}(i)
	}

	wg.Wait()
	close(results)

	byID := make(map[int]handler.Task)

	for rec := range results {
		requireStatus(t, rec, http.StatusCreated)
		task := decodeTask(t, rec)

		if task.Done {
			t.Fatal("созданная задача имеет done=true")
		}
		if _, exists := byID[task.Id]; exists {
			t.Fatalf("повторный ID: %d", task.Id)
		}
		byID[task.Id] = task
	}

	tasks := listTasks(t, api)
	if len(tasks) != count {
		t.Fatalf("потеря задач: получили %d, хотели %d", len(tasks), count)
	}

	for _, task := range tasks {
		if task != byID[task.Id] {
			t.Fatalf("задача не совпадает с ответом создания: %+v", task)
		}
	}
}

// Одновременно читаем одну задачу, список и меняем эту же задачу.
// Именно здесь -race может заметить забытое копирование под мьютексом.
func TestConcurrentReadAndUpdate(t *testing.T) {
	api := newTestAPI()
	created := createTask(t, api, "Исходная")
	path := fmt.Sprintf("/tasks/%d", created.Id)

	const workers = 12
	const iterations = 40

	results := make(chan *httptest.ResponseRecorder, workers*iterations)
	var wg sync.WaitGroup

	for worker := 0; worker < workers; worker++ {
		wg.Add(1)
		go func(worker int) {
			defer wg.Done()

			for i := 0; i < iterations; i++ {
				switch worker % 3 {
				case 0:
					body := fmt.Sprintf(
						`{"title":"Worker %d step %d","done":%t}`,
						worker, i, i%2 == 0,
					)
					results <- request(api, http.MethodPut, path, body)
				case 1:
					results <- request(api, http.MethodGet, path, "")
				case 2:
					results <- request(api, http.MethodGet, "/tasks/", "")
				}
			}
		}(worker)
	}

	wg.Wait()
	close(results)

	for rec := range results {
		requireStatus(t, rec, http.StatusOK)
		requireJSON(t, rec)

		if strings.HasPrefix(strings.TrimSpace(rec.Body.String()), "[") {
			var tasks []handler.Task
			if err := json.Unmarshal(rec.Body.Bytes(), &tasks); err != nil {
				t.Fatal(err)
			}
			if len(tasks) != 1 || tasks[0].Id != created.Id {
				t.Fatalf("неверный список: %+v", tasks)
			}
		} else {
			task := decodeTask(t, rec)
			if task.Id != created.Id || task.Title == "" {
				t.Fatalf("неверная задача: %+v", task)
			}
		}
	}
}

// Изменяем разные задачи параллельно и проверяем каждое изменение.
func TestConcurrentUpdateDifferentTasks(t *testing.T) {
	api := newTestAPI()
	const count = 40

	tasks := make([]handler.Task, 0, count)
	for i := 0; i < count; i++ {
		tasks = append(tasks, createTask(t, api, fmt.Sprintf("Task %d", i)))
	}

	results := make(chan *httptest.ResponseRecorder, count)
	var wg sync.WaitGroup

	for _, task := range tasks {
		wg.Add(1)
		go func(task handler.Task) {
			defer wg.Done()

			body := fmt.Sprintf(
				`{"title":"Updated %d","done":true}`,
				task.Id,
			)
			results <- request(api, http.MethodPut,
				fmt.Sprintf("/tasks/%d", task.Id), body)
		}(task)
	}

	wg.Wait()
	close(results)

	for rec := range results {
		requireStatus(t, rec, http.StatusOK)
		decodeTask(t, rec)
	}

	for _, task := range tasks {
		got := getTask(t, api, task.Id)
		if got.Title != fmt.Sprintf("Updated %d", task.Id) || !got.Done {
			t.Fatalf("изменение потерялось: %+v", got)
		}
	}
}

// Несколько запросов пытаются удалить одну задачу.
// Ровно один должен получить 204, остальные — 404.
func TestConcurrentDeleteSameTask(t *testing.T) {
	api := newTestAPI()
	task := createTask(t, api, "Удалить")
	path := fmt.Sprintf("/tasks/%d", task.Id)

	const count = 30
	results := make(chan *httptest.ResponseRecorder, count)
	var wg sync.WaitGroup

	for i := 0; i < count; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			results <- request(api, http.MethodDelete, path, "")
		}()
	}

	wg.Wait()
	close(results)

	deleted := 0
	for rec := range results {
		switch rec.Code {
		case http.StatusNoContent:
			deleted++
			if rec.Body.Len() != 0 {
				t.Fatal("204 содержит тело")
			}
		case http.StatusNotFound:
			requireError(t, rec, http.StatusNotFound)
		default:
			t.Fatalf("неожиданный статус: %d; %s", rec.Code, rec.Body.String())
		}
	}

	if deleted != 1 {
		t.Fatalf("успешных удалений: %d, ожидали 1", deleted)
	}
}

// Создание, чтение, обновление и удаление выполняются параллельно.
func TestConcurrentMixedOperations(t *testing.T) {
	api := newTestAPI()
	const workers = 16
	const iterations = 15

	errs := make(chan error, workers)
	var wg sync.WaitGroup

	for worker := 0; worker < workers; worker++ {
		wg.Add(1)
		go func(worker int) {
			defer wg.Done()

			for i := 0; i < iterations; i++ {
				body := fmt.Sprintf(`{"title":"Task %d-%d"}`, worker, i)
				rec := request(api, http.MethodPost, "/tasks/", body)
				if rec.Code != http.StatusCreated {
					errs <- fmt.Errorf("POST: %d, %s", rec.Code, rec.Body.String())
					return
				}

				var task handler.Task
				if err := json.Unmarshal(rec.Body.Bytes(), &task); err != nil {
					errs <- err
					return
				}

				path := fmt.Sprintf("/tasks/%d", task.Id)

				rec = request(api, http.MethodGet, "/tasks/", "")
				if rec.Code != http.StatusOK || !json.Valid(rec.Body.Bytes()) {
					errs <- fmt.Errorf("GET list: %d, %s", rec.Code, rec.Body.String())
					return
				}

				rec = request(api, http.MethodPut, path,
					`{"title":"Updated","done":true}`)
				if rec.Code != http.StatusOK {
					errs <- fmt.Errorf("PUT: %d, %s", rec.Code, rec.Body.String())
					return
				}

				rec = request(api, http.MethodGet, path, "")
				var updated handler.Task
				err := json.Unmarshal(rec.Body.Bytes(), &updated)
				if rec.Code != http.StatusOK || err != nil ||
					updated.Id != task.Id || updated.Title != "Updated" || !updated.Done {
					errs <- fmt.Errorf("GET task: %d, %s", rec.Code, rec.Body.String())
					return
				}

				rec = request(api, http.MethodDelete, path, "")
				if rec.Code != http.StatusNoContent || rec.Body.Len() != 0 {
					errs <- fmt.Errorf("DELETE: %d, %s", rec.Code, rec.Body.String())
					return
				}
			}
		}(worker)
	}

	wg.Wait()
	close(errs)

	for err := range errs {
		t.Error(err)
	}

	if tasks := listTasks(t, api); len(tasks) != 0 {
		t.Fatalf("после удаления остались задачи: %d", len(tasks))
	}
}
