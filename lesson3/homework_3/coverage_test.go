package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const (
	testAccessToken = "test-token"
	testDatasetPath = "dataset.xml"
)

func sendRequestTest(t *testing.T, target, token string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(http.MethodGet, target, nil)
	if token != "" {
		req.Header.Set("AccessToken", token)
	}

	rec := httptest.NewRecorder()
	SearchServer(rec, req)
	return rec
}

func decodeUsers(t *testing.T, rec *httptest.ResponseRecorder) []User {
	t.Helper()

	var got []User
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("не удалось декодировать JSON-массив пользователей: %v; status=%d; body=%q",
			err, rec.Code, rec.Body.String())
	}
	return got
}

func requireStatus(t *testing.T, rec *httptest.ResponseRecorder, want int) {
	t.Helper()

	if rec.Code != want {
		t.Fatalf("неверный HTTP-статус: ожидался %d, получен %d; body=%q",
			want, rec.Code, rec.Body.String())
	}

	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Errorf("неверный Content-Type: ожидался %q, получен %q",
			"application/json", got)
	}
}

func requireUsers(t *testing.T, got []User, wantIDs []int, wantNames []string) {
	t.Helper()

	if len(got) != len(wantIDs) || len(got) != len(wantNames) {
		t.Fatalf("неверное количество пользователей: ожидалось %d, получено %d; полученные записи: %#v",
			len(wantIDs), len(got), got)
	}

	for i := range got {
		if got[i].ID != wantIDs[i] || got[i].Name != wantNames[i] {
			t.Errorf("пользователь [%d]: ожидались ID=%d Name=%q, получены ID=%d Name=%q",
				i, wantIDs[i], wantNames[i], got[i].ID, got[i].Name)
		}
		if got[i].About == "" {
			t.Errorf("у пользователя ID=%d неожиданно пустое поле About", got[i].ID)
		}
	}
}

func logUsers(t *testing.T, users []User) {
	t.Helper()

	t.Logf("получено пользователей: %d", len(users))
	for _, user := range users {
		t.Logf("  ID=%d | Name=%q | Age=%d | Gender=%q | About: %d символов",
			user.ID, user.Name, user.Age, user.Gender, len(user.About))
	}
}

func requireErrorResponse(t *testing.T, rec *httptest.ResponseRecorder, wantStatus int, wantMessage string) {
	t.Helper()

	requireStatus(t, rec, wantStatus)

	var got SearchErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("не удалось декодировать JSON-ошибку: %v; body=%q",
			err, rec.Body.String())
	}

	if got.Error != wantMessage {
		t.Fatalf("неверное сообщение ошибки: ожидалось %q, получено %q",
			wantMessage, got.Error)
	}

	t.Logf("HTTP %d | Error=%q", rec.Code, got.Error)
}

func requireSortedAges(t *testing.T, users []User, ascending bool) {
	t.Helper()

	for i := 1; i < len(users); i++ {
		if ascending && users[i-1].Age > users[i].Age {
			t.Fatalf("возраст не отсортирован по возрастанию на индексах %d и %d: %d > %d",
				i-1, i, users[i-1].Age, users[i].Age)
		}
		if !ascending && users[i-1].Age < users[i].Age {
			t.Fatalf("возраст не отсортирован по убыванию на индексах %d и %d: %d < %d",
				i-1, i, users[i-1].Age, users[i].Age)
		}
	}
}

func setDatasetPathForTest(t *testing.T, path string) {
	t.Helper()

	previousPath := pathToFile
	pathToFile = path

	t.Cleanup(func() {
		pathToFile = previousPath
	})
}

func TestSearchServer_SuccessfulRequests(t *testing.T) {
	setDatasetPathForTest(t, testDatasetPath)

	t.Run("default field sorts names ascending and applies limit", func(t *testing.T) {
		rec := sendRequestTest(t, "/?limit=2&offset=0&order_by=1", testAccessToken)
		requireStatus(t, rec, http.StatusOK)

		got := decodeUsers(t, rec)
		logUsers(t, got)
		requireUsers(t, got, []int{15, 16}, []string{"Allison Valdez", "Annie Osborn"})
	})

	t.Run("names descending", func(t *testing.T) {
		rec := sendRequestTest(t, "/?limit=2&offset=0&order_by=-1", testAccessToken)
		requireStatus(t, rec, http.StatusOK)

		got := decodeUsers(t, rec)
		logUsers(t, got)
		requireUsers(t, got, []int{13, 33}, []string{"Whitley Davidson", "Twila Snow"})
	})

	t.Run("search by name", func(t *testing.T) {
		rec := sendRequestTest(t, "/?limit=5&offset=0&query=Boyd&order_by=1", testAccessToken)
		requireStatus(t, rec, http.StatusOK)

		got := decodeUsers(t, rec)
		logUsers(t, got)
		requireUsers(t, got, []int{0}, []string{"Boyd Wolf"})

		if got[0].Age != 22 || got[0].Gender != "male" {
			t.Errorf("поля пользователя декодированы неверно: Age=%d Gender=%q",
				got[0].Age, got[0].Gender)
		}
	})

	t.Run("search in About", func(t *testing.T) {
		phrase := "Nulla cillum enim voluptate consequat laborum"
		target := "/?limit=5&offset=0&query=" + url.QueryEscape(phrase) + "&order_by=1"

		rec := sendRequestTest(t, target, testAccessToken)
		requireStatus(t, rec, http.StatusOK)

		got := decodeUsers(t, rec)
		logUsers(t, got)
		requireUsers(t, got, []int{0}, []string{"Boyd Wolf"})

		if !strings.Contains(got[0].About, phrase) {
			t.Errorf("поле About пользователя не содержит искомую подстроку %q", phrase)
		}
	})

	t.Run("query without matches returns empty result", func(t *testing.T) {
		rec := sendRequestTest(t, "/?limit=5&offset=0&query=NoSuchNeedle_XYZ_123&order_by=1", testAccessToken)
		requireStatus(t, rec, http.StatusOK)

		got := decodeUsers(t, rec)
		logUsers(t, got)

		if len(got) != 0 {
			t.Fatalf("ожидался пустой результат, получено %d пользователей", len(got))
		}
	})

	t.Run("IDs ascending", func(t *testing.T) {
		rec := sendRequestTest(t, "/?limit=3&offset=0&order_field=Id&order_by=1", testAccessToken)
		requireStatus(t, rec, http.StatusOK)

		got := decodeUsers(t, rec)
		logUsers(t, got)
		requireUsers(t, got, []int{0, 1, 2}, []string{"Boyd Wolf", "Hilda Mayer", "Brooks Aguilar"})
	})

	t.Run("IDs descending", func(t *testing.T) {
		rec := sendRequestTest(t, "/?limit=3&offset=0&order_field=Id&order_by=-1", testAccessToken)
		requireStatus(t, rec, http.StatusOK)

		got := decodeUsers(t, rec)
		logUsers(t, got)
		requireUsers(t, got, []int{34, 33, 32}, []string{"Kane Sharp", "Twila Snow", "Christy Knapp"})
	})

	t.Run("ages ascending", func(t *testing.T) {
		rec := sendRequestTest(t, "/?limit=50&offset=0&order_field=Age&order_by=1", testAccessToken)
		requireStatus(t, rec, http.StatusOK)

		got := decodeUsers(t, rec)
		logUsers(t, got)

		if len(got) != 35 {
			t.Fatalf("ожидалось 35 записей из dataset.xml, получено %d", len(got))
		}
		requireSortedAges(t, got, true)
	})

	t.Run("ages descending", func(t *testing.T) {
		rec := sendRequestTest(t, "/?limit=50&offset=0&order_field=Age&order_by=-1", testAccessToken)
		requireStatus(t, rec, http.StatusOK)

		got := decodeUsers(t, rec)
		logUsers(t, got)

		if len(got) != 35 {
			t.Fatalf("ожидалось 35 записей из dataset.xml, получено %d", len(got))
		}
		requireSortedAges(t, got, false)
	})

	t.Run("order_by zero preserves source order", func(t *testing.T) {
		rec := sendRequestTest(t, "/?limit=3&offset=0&order_field=Name&order_by=0", testAccessToken)
		requireStatus(t, rec, http.StatusOK)

		got := decodeUsers(t, rec)
		logUsers(t, got)
		requireUsers(t, got, []int{0, 1, 2}, []string{"Boyd Wolf", "Hilda Mayer", "Brooks Aguilar"})
	})

	t.Run("offset applies when order_by is zero", func(t *testing.T) {
		rec := sendRequestTest(t, "/?limit=2&offset=3&order_field=Name&order_by=0", testAccessToken)
		requireStatus(t, rec, http.StatusOK)

		got := decodeUsers(t, rec)
		logUsers(t, got)
		requireUsers(t, got, []int{3, 4}, []string{"Everett Dillard", "Owen Lynn"})
	})

	t.Run("offset beyond dataset returns empty result", func(t *testing.T) {
		rec := sendRequestTest(t, "/?limit=2&offset=100&order_by=0", testAccessToken)
		requireStatus(t, rec, http.StatusOK)

		got := decodeUsers(t, rec)
		logUsers(t, got)

		if len(got) != 0 {
			t.Fatalf("ожидался пустой результат при offset=100, получено %d пользователей", len(got))
		}
	})

	t.Run("limit greater than dataset size returns all users", func(t *testing.T) {
		rec := sendRequestTest(t, "/?limit=50&offset=0&order_by=0", testAccessToken)
		requireStatus(t, rec, http.StatusOK)

		got := decodeUsers(t, rec)
		logUsers(t, got)

		if len(got) != 35 {
			t.Fatalf("ожидалось 35 пользователей, получено %d", len(got))
		}
	})

	t.Run("offset is applied after query filtering", func(t *testing.T) {
		rec := sendRequestTest(t, "/?limit=2&offset=1&query=Boyd&order_by=1", testAccessToken)
		requireStatus(t, rec, http.StatusOK)

		got := decodeUsers(t, rec)
		logUsers(t, got)

		if len(got) != 0 {
			t.Fatalf("после фильтрации одного пользователя и offset=1 ожидался пустой результат, получено %d", len(got))
		}
	})

	t.Run("offset applies to descending ID order", func(t *testing.T) {
		rec := sendRequestTest(t, "/?limit=2&offset=1&order_field=Id&order_by=-1", testAccessToken)
		requireStatus(t, rec, http.StatusOK)

		got := decodeUsers(t, rec)
		logUsers(t, got)
		requireUsers(t, got, []int{33, 32}, []string{"Twila Snow", "Christy Knapp"})
	})
}

func TestSearchServer_RequestErrors(t *testing.T) {
	setDatasetPathForTest(t, testDatasetPath)

	tests := []struct {
		name        string
		target      string
		token       string
		wantStatus  int
		wantMessage string
	}{
		{
			name: "missing AccessToken", target: "/?limit=2&offset=0&order_by=1", token: "",
			wantStatus: http.StatusUnauthorized, wantMessage: "bad AccessToken",
		},
		{
			name: "limit is not an integer", target: "/?limit=abc&offset=0&order_by=1", token: testAccessToken,
			wantStatus: http.StatusBadRequest, wantMessage: "Параметр limit не получилось перевести в число",
		},
		{
			name: "limit is missing", target: "/?offset=0&order_by=1", token: testAccessToken,
			wantStatus: http.StatusBadRequest, wantMessage: "Параметр limit не получилось перевести в число",
		},
		{
			name: "limit is zero", target: "/?limit=0&offset=0&order_by=1", token: testAccessToken,
			wantStatus: http.StatusBadRequest, wantMessage: "Параметр limit должен быть положительным",
		},
		{
			name: "limit is negative", target: "/?limit=-2&offset=0&order_by=1", token: testAccessToken,
			wantStatus: http.StatusBadRequest, wantMessage: "Параметр limit должен быть положительным",
		},
		{
			name: "offset is not an integer", target: "/?limit=2&offset=abc&order_by=1", token: testAccessToken,
			wantStatus: http.StatusBadRequest, wantMessage: "Параметр offset не получилось перевести в число",
		},
		{
			name: "offset is missing", target: "/?limit=2&order_by=1", token: testAccessToken,
			wantStatus: http.StatusBadRequest, wantMessage: "Параметр offset не получилось перевести в число",
		},
		{
			name: "offset is negative", target: "/?limit=2&offset=-1&order_by=1", token: testAccessToken,
			wantStatus: http.StatusBadRequest, wantMessage: "Параметр offset меньше нуля",
		},
		{
			name: "order_by is not an integer", target: "/?limit=2&offset=0&order_by=abc", token: testAccessToken,
			wantStatus: http.StatusBadRequest, wantMessage: "Параметр order_by не получилось перевести в число",
		},
		{
			name: "order_by is missing", target: "/?limit=2&offset=0", token: testAccessToken,
			wantStatus: http.StatusBadRequest, wantMessage: "Параметр order_by не получилось перевести в число",
		},
		{
			name: "order_by is outside allowed values", target: "/?limit=2&offset=0&order_by=5", token: testAccessToken,
			wantStatus: http.StatusBadRequest, wantMessage: "Поле order_by не равно '1', '-1' или '0'",
		},
		{
			name: "unsupported order_field", target: "/?limit=2&offset=0&order_field=Height&order_by=1", token: testAccessToken,
			wantStatus: http.StatusBadRequest, wantMessage: ErrorBadOrderField,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rec := sendRequestTest(t, tc.target, tc.token)
			requireErrorResponse(t, rec, tc.wantStatus, tc.wantMessage)
		})
	}
}

func TestSearchServer_FileErrors(t *testing.T) {
	t.Run("file does not exist", func(t *testing.T) {
		missingPath := filepath.Join(t.TempDir(), "missing-dataset.xml")
		setDatasetPathForTest(t, missingPath)

		rec := sendRequestTest(t, "/?limit=2&offset=0&order_by=1", testAccessToken)
		requireErrorResponse(t, rec, http.StatusInternalServerError, "Internal Server Error")
	})

	t.Run("XML cannot be decoded", func(t *testing.T) {
		brokenPath := filepath.Join(t.TempDir(), "broken-dataset.xml")
		brokenXML := []byte(`<root><row><id>not-an-integer</id></row></root>`)

		if err := os.WriteFile(brokenPath, brokenXML, 0o600); err != nil {
			t.Fatalf("не удалось создать временный XML-файл: %v", err)
		}

		setDatasetPathForTest(t, brokenPath)

		rec := sendRequestTest(t, "/?limit=2&offset=0&order_by=1", testAccessToken)
		requireErrorResponse(t, rec, http.StatusInternalServerError, "Internal Server Error")
	})
}
