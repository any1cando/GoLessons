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

func openFile(file *os.File) Dataset {
	var dataset Dataset
	xml.NewDecoder(file).Decode(&dataset)
	return dataset
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

// TODO: Подумать еще на проверку того, а как вообще лимит+1 учитывать

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
	if orderField != "Id" && orderField != "Age" && orderField != "Name" || offset < 0 {
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

func SearchServer(w http.ResponseWriter, r *http.Request) {
	var finalListOfUsers []User
	paramsQuery := r.URL.Query()

	// limit - максимальное количество записей, которое мы можем вернуть (+1, чтобы клиент у себя сортил флаг NextPage - true/false
	limit, err := strconv.Atoi(paramsQuery.Get("limit"))
	if err != nil {
		fmt.Fprintln(w, "Не получилось конвертировать параметр limit в число")
		return
	}
	if limit <= 0 {
		fmt.Fprintln(w, "Лимит пользователей 0! Сервер не может вернуть ни одного пользователя!")
		return
	}

	file, _ := os.Open("dataset.xml")
	defer file.Close()
	dataset := openFile(file)

	query := paramsQuery.Get("query")                      // текст (подстрока), который нужно найти в нашем файле dataset.xml в полях Name/About
	offset, err := strconv.Atoi(paramsQuery.Get("offset")) // позволяет получить список с определенного индекса в уже полученном отсортированном слайсе
	if err != nil {
		fmt.Fprintln(w, "Не получилось конвертировать параметр offset в число")
		return
	}

	order_field := paramsQuery.Get("order_field")              // указывает по какому полю сортировать собранный слайс подходящих юзеров
	order_by, err := strconv.Atoi(paramsQuery.Get("order_by")) // задает направление сортировки: 1 по возрастанию, 0 как встретилось, -1 по убыванию
	if err != nil {
		fmt.Fprintln(w, "Не получилось конвертировать параметр order_by в число")
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
		fmt.Fprintln(w, "Сортировка невозможна: по параметру order_field нет сортировки!")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(sortedFinalListOfUsers)
	if err != nil {
		fmt.Println("Ошибка: не получилось собрать ответ в JSON")
		return
	}
	// searcherParams.Add("limit", strconv.Itoa(req.Limit))
	//	searcherParams.Add("offset", strconv.Itoa(req.Offset))
	//	searcherParams.Add("query", req.Query)
	//	searcherParams.Add("order_field", req.OrderField)
	//	searcherParams.Add("order_by", strconv.Itoa(req.OrderBy))

}
