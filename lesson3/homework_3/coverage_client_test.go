package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestFindUsersNegativeLimit(t *testing.T) {
	searchClient := SearchClient{}

	result, err := searchClient.FindUsers(SearchRequest{Limit: -1})

	if err == nil {
		t.Fatal("Ожидалась ошибка при отрицательном Limit")
	}
	if err.Error() != "limit must be > 0" {
		t.Errorf("Неожиданная ошибка: %v", err)
	}
	if result != nil {
		t.Errorf("Ожидался nil вместо результата, получено: %+v", result)
	}
}

func TestFindUsersNegativeOffset(t *testing.T) {
	searchClient := SearchClient{}

	result, err := searchClient.FindUsers(SearchRequest{
		Limit:  5,
		Offset: -1,
	})

	if err == nil {
		t.Fatal("Ожидалась ошибка при отрицательном Offset")
	}
	if err.Error() != "offset must be > 0" {
		t.Errorf("Неожиданная ошибка: %v", err)
	}
	if result != nil {
		t.Errorf("Ожидался nil вместо результата, получено: %+v", result)
	}
}

// Проверяем параметры запроса, пользователей и NextPage.
func TestFindUsersSuccessResponses(t *testing.T) {
	alice := User{ID: 1, Name: "Alice Smith", Age: 25, About: "Developer", Gender: "female"}
	ben := User{ID: 2, Name: "Ben Wolf", Age: 30, About: "Designer", Gender: "male"}
	charles := User{ID: 3, Name: "Charles Brown", Age: 35, About: "Teacher", Gender: "male"}
	many := make([]User, 26)
	for i := range many {
		many[i] = User{ID: i, Name: fmt.Sprintf("User %02d", i), Age: 20 + i}
	}
	cases := []struct {
		name        string
		limit       int
		sentLimit   string
		serverUsers []User
		wantUsers   []User
		nextPage    bool
	}{
		{"last_page", 3, "4", []User{alice, ben}, []User{alice, ben}, false},
		{"next_page", 2, "3", []User{alice, ben, charles}, []User{alice, ben}, true},
		{"exact_limit", 2, "3", []User{alice, ben}, []User{alice, ben}, false},
		{"empty_result", 2, "3", []User{}, []User{}, false},
		{"limit_capped", 100, "26", many, many[:25], true},
		{"zero_limit", 0, "1", []User{alice}, []User{}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := SearchRequest{Limit: tc.limit, Offset: 4, Query: "Alice & Ben+текст", OrderField: "Name", OrderBy: OrderByDesc}
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					t.Errorf("Метод: %q, ожидали GET", r.Method)
				}
				if token := r.Header.Get("AccessToken"); token != "test-token" {
					t.Errorf("Неверный AccessToken: %q", token)
				}
				want := url.Values{
					"limit": {tc.sentLimit}, "offset": {"4"}, "query": {req.Query},
					"order_field": {"Name"}, "order_by": {"-1"},
				}
				if got := r.URL.Query(); !reflect.DeepEqual(got, want) {
					t.Errorf("Параметры: %v, ожидали %v", got, want)
				}
				w.Header().Set("Content-Type", "application/json")
				if err := json.NewEncoder(w).Encode(tc.serverUsers); err != nil {
					t.Errorf("Ошибка записи ответа: %v", err)
				}
			}))
			defer server.Close()
			searchClient := SearchClient{URL: server.URL, AccessToken: "test-token"}
			result, err := searchClient.FindUsers(req)
			if err != nil {
				t.Fatalf("Неожиданная ошибка: %v", err)
			}
			if result == nil {
				t.Fatal("Получен nil вместо результата")
			}
			if !reflect.DeepEqual(result.Users, tc.wantUsers) {
				t.Errorf("Users: %+v, ожидали %+v", result.Users, tc.wantUsers)
			}
			if result.NextPage != tc.nextPage {
				t.Errorf("NextPage: %v, ожидали %v", result.NextPage, tc.nextPage)
			}
		})
	}
}

// Тестовый сервер намеренно возвращает ошибки и повреждённый JSON.
func TestFindUsersErrorResponses(t *testing.T) {
	cases := []struct {
		name      string
		status    int
		body      string
		wantError string
		prefix    bool
	}{
		{"unauthorized", 401, "", "bad AccessToken", false},
		{"server_error", 500, "", "SearchServer fatal error", false},
		{"bad_order_field", 400, `{"Error":"OrderField invalid"}`, "OrderFeld Unknown invalid", false},
		{"other_bad_request", 400, `{"Error":"invalid offset"}`, "unknown bad request error: invalid offset", false},
		{"broken_error_json", 400, `{`, "cant unpack error json: ", true},
		{"broken_result_json", 200, `{`, "cant unpack result json: ", true},
		{"wrong_result_type", 200, `{"Name":"Alice"}`, "cant unpack result json: ", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tc.status)
				if _, err := fmt.Fprint(w, tc.body); err != nil {
					t.Errorf("Ошибка записи: %v", err)
				}
			}))
			defer server.Close()
			searchClient := SearchClient{URL: server.URL, AccessToken: "test-token"}
			result, err := searchClient.FindUsers(SearchRequest{Limit: 2, OrderField: "Unknown"})
			if err == nil {
				t.Fatal("Ожидалась ошибка")
			}
			if tc.prefix {
				if !strings.HasPrefix(err.Error(), tc.wantError) {
					t.Errorf("Ожидался префикс %q, получено: %v", tc.wantError, err)
				}
			} else if err.Error() != tc.wantError {
				t.Errorf("Ожидалась ошибка %q, получено: %v", tc.wantError, err)
			}
			if result != nil {
				t.Errorf("Ожидался nil, получено: %+v", result)
			}
		})
	}
}

func TestFindUsersRequestTimeout(t *testing.T) {
	// Глобальный клиент восстанавливается после теста. Не использовать t.Parallel().
	oldClient := client
	client = &http.Client{Timeout: 50 * time.Millisecond}
	t.Cleanup(func() { client = oldClient })
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Ждём отмены запроса клиентом вместо отправки ответа.
		<-r.Context().Done()
	}))
	defer server.Close()
	searchClient := SearchClient{URL: server.URL}
	result, err := searchClient.FindUsers(SearchRequest{Limit: 1})
	if err == nil {
		t.Fatal("Ожидалась ошибка таймаута")
	}
	want := "timeout for limit=2&offset=0&order_by=0&order_field=&query="
	if err.Error() != want {
		t.Errorf("Ожидалось %q, получено: %v", want, err)
	}
	if result != nil {
		t.Errorf("Ожидался nil, получено: %+v", result)
	}
}

func TestFindUsersUnsupportedProtocol(t *testing.T) {
	// Валидный URL с протоколом, который HTTP-клиент не поддерживает.
	searchClient := SearchClient{URL: "ftp://example.com"}
	result, err := searchClient.FindUsers(SearchRequest{Limit: 1})
	if err == nil {
		t.Fatal("Ожидалась ошибка неподдерживаемого протокола")
	}
	if !strings.HasPrefix(err.Error(), "unknown error ") ||
		!strings.Contains(err.Error(), `unsupported protocol scheme "ftp"`) {
		t.Errorf("Неожиданная ошибка: %v", err)
	}
	if result != nil {
		t.Errorf("Ожидался nil, получено: %+v", result)
	}
}
