package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
)


var pathToFile = "dataset.xml"

type UserXML struct {
	ID        int    `xml:"id"`
	FirstName string `xml:"first_name"`
	LastName  string `xml:"last_name"`
	Age       int    `xml:"age"`
	About     string `xml:"about"`
	Gender    string `xml:"gender"`
} // TODO: Доделать HTTP статусы

type Dataset struct {
	Users []UserXML `xml:"row"`
}

func openFile(file *os.File) (Dataset, error) {
	var dataset Dataset
	err := xml.NewDecoder(file).Decode(&dataset)
	if err != nil {
		fmt.Println("Не получилось преобразовать XML в структуры")
		return dataset, fmt.Errorf("can't transform xml to code")
	}
	return dataset, nil
}

func fillUsersSliceWithQuery(dataset Dataset, query string) []User {
	var usersWeNeed []User

	for index := range dataset.Users {

		fullNameOfUser := dataset.Users[index].FirstName + " " + dataset.Users[index].LastName
		if strings.Contains(fullNameOfUser, query) || strings.Contains(dataset.Users[index].About, query) {
			usersWeNeed = append(
				usersWeNeed,
				User{
					ID:     dataset.Users[index].ID,
					Name:   fullNameOfUser,
					Age:    dataset.Users[index].Age,
					About:  dataset.Users[index].About,
					Gender: dataset.Users[index].Gender,
				})
		}

	}
	return usersWeNeed
}

func fillAllUsers(dataset Dataset) []User {
	var usersWeNeed []User
	for index := range dataset.Users {

		fullNameOfUser := dataset.Users[index].FirstName + " " + dataset.Users[index].LastName
		usersWeNeed = append(
			usersWeNeed,
			User{
				ID:     dataset.Users[index].ID,
				Name:   fullNameOfUser,
				Age:    dataset.Users[index].Age,
				About:  dataset.Users[index].About,
				Gender: dataset.Users[index].Gender,
			})

	}
	return usersWeNeed
}

func sortFinalList(listWeNeedToSort []User, orderField string, orderBy int, limit int, offset int) ([]User, error) {
	if orderField == "" {
		orderField = "Name"
	}
	if orderField != "Id" && orderField != "Age" && orderField != "Name" {
		fmt.Println("Нельзя отсортировать такой список")
		return listWeNeedToSort, fmt.Errorf("can't be sorted")
	}

	switch orderField {
	case "Name":
		switch orderBy {
		case 1:
			sort.Slice(listWeNeedToSort, func(i, j int) bool {
				return listWeNeedToSort[i].Name < listWeNeedToSort[j].Name
			})
		case -1:
			sort.Slice(listWeNeedToSort, func(i, j int) bool {
				return listWeNeedToSort[i].Name > listWeNeedToSort[j].Name
			})
		case 0:
		default:
			return listWeNeedToSort, fmt.Errorf("invalid order_by field")
		}
	case "Age":
		switch orderBy {
		case 1:
			sort.Slice(listWeNeedToSort, func(i, j int) bool {
				return listWeNeedToSort[i].Age < listWeNeedToSort[j].Age
			})
		case -1:
			sort.Slice(listWeNeedToSort, func(i, j int) bool {
				return listWeNeedToSort[i].Age > listWeNeedToSort[j].Age
			})
		case 0:
		default:
			return listWeNeedToSort, fmt.Errorf("invalid order_by field")
		}
	case "Id":
		switch orderBy {
		case 1:
			sort.Slice(listWeNeedToSort, func(i, j int) bool {
				return listWeNeedToSort[i].ID < listWeNeedToSort[j].ID
			})
		case -1:
			sort.Slice(listWeNeedToSort, func(i, j int) bool {
				return listWeNeedToSort[i].ID > listWeNeedToSort[j].ID
			})
		case 0:
		default:
			return listWeNeedToSort, fmt.Errorf("invalid order_by field")
		}
	}

	// Теперь отбрасываем лишние данные, отсекая все данные, которые были до offset, делая проверку на границы
	if offset >= len(listWeNeedToSort) {
		return []User{}, nil // возвращаем пустой список, если мы срезаем все offset = list(len)
	}
	listWeNeedToSort = listWeNeedToSort[offset:]

	// Теперь отбрасываем лишние данные, отсекая данные до количества limit. Если у нас, например, 4 элемента подходит, а limit 6, то возвращаем просто все
	if len(listWeNeedToSort) <= limit {
		return listWeNeedToSort, nil
	}

	// А если у нас 10 элементов подходит, но вернуть мы должны только 5 - то возвращаем первые 5 с индексами (0, 1, 2, 3, 4) - limit, правая граница, здесь будет учитываться с "-1"
	return listWeNeedToSort[0:limit], nil

}

func setErrorStatus(w http.ResponseWriter, status int, textError string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	errResp := SearchErrorResponse{Error: textError}
	err := json.NewEncoder(w).Encode(errResp)
	if err != nil {
		fmt.Println("Не получилось сформировать json файл")
		return
	}
}

func SearchServer(w http.ResponseWriter, r *http.Request) {

	accessToken := r.Header.Get("AccessToken")
	if accessToken == "" {
		setErrorStatus(w, http.StatusUnauthorized, "bad AccessToken")
		return
	}

	var finalListOfUsers []User
	paramsQuery := r.URL.Query()

	// limit - максимальное количество записей, которое мы можем вернуть (+1, чтобы клиент у себя сортил флаг NextPage - true/false
	limit, err := strconv.Atoi(paramsQuery.Get("limit"))
	if err != nil {
		setErrorStatus(w, http.StatusBadRequest, "Параметр limit не получилось перевести в число")
		return
	}
	if limit <= 0 {
		setErrorStatus(w, http.StatusBadRequest, "Параметр limit должен быть положительным")
		return
	}

	file, err := os.Open(pathToFile)
	if err != nil {
		setErrorStatus(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}
	defer file.Close()
	dataset, err := openFile(file)
	if err != nil {
		setErrorStatus(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	query := paramsQuery.Get("query")                      // текст (подстрока), который нужно найти в нашем файле dataset.xml в полях Name/About
	offset, err := strconv.Atoi(paramsQuery.Get("offset")) // позволяет получить список с определенного индекса в уже полученном отсортированном слайсе
	if err != nil {
		// Запаковали ошибку 400
		setErrorStatus(w, http.StatusBadRequest, "Параметр offset не получилось перевести в число")
		return
	}

	if offset < 0 {
		setErrorStatus(w, http.StatusBadRequest, "Параметр offset меньше нуля")
		return
	}

	order_field := paramsQuery.Get("order_field")              // указывает по какому полю сортировать собранный слайс подходящих юзеров
	order_by, err := strconv.Atoi(paramsQuery.Get("order_by")) // задает направление сортировки: 1 по возрастанию, 0 как встретилось, -1 по убыванию
	if err != nil {
		// Запаковали ошибку 400
		setErrorStatus(w, http.StatusBadRequest, "Параметр order_by не получилось перевести в число")
		return
	}

	if query == "" {
		fmt.Println("Пустая строка запроса: возвращаем всех пользователей")
		finalListOfUsers = fillAllUsers(dataset)
	} else {
		fmt.Println("Строка запроса что-то содержит, ищем схожие подстроки")
		finalListOfUsers = fillUsersSliceWithQuery(dataset, query)
	}

	fmt.Println("Сортируем полученный список с помощью order_field + order_by. Также убираем лишние данные по limit")
	sortedFinalListOfUsers, err := sortFinalList(finalListOfUsers, order_field, order_by, limit, offset)
	if err != nil && err.Error() == "can't be sorted" {
		// Запаковали ошибку 400
		setErrorStatus(w, http.StatusBadRequest, ErrorBadOrderField)
		return
	}
	if err != nil && err.Error() == "invalid order_by field" {
		setErrorStatus(w, http.StatusBadRequest, "Поле order_by не равно '1', '-1' или '0'")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(sortedFinalListOfUsers)
	if err != nil {
		fmt.Println("Не получилось сформировать json файл")
		return
	}

	// searcherParams.Add("limit", strconv.Itoa(req.Limit))
	//	searcherParams.Add("offset", strconv.Itoa(req.Offset))
	//	searcherParams.Add("query", req.Query)
	//	searcherParams.Add("order_field", req.OrderField)
	//	searcherParams.Add("order_by", strconv.Itoa(req.OrderBy))

}
