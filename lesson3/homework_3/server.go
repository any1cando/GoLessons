package main

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"os"
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
} // TODO: Разобрать с нейронкой, можно ли не создавать UserXML, а сразу напрямую работать со структурой User?

type Dataset struct {
	Users []UserXML `xml:"row"`
}

func fillUsersSliceWithQuery(limit int, query string, dataset Dataset, usersWeNeed []User) []User {
	var counter int // создали счетчик, куда будем делать +1, если нашли подходящего пользователя (для проверки параметра limit)

	for index := range dataset.Users {
		if counter == limit {
			break
		}
		fullNameOfUser := dataset.Users[index].FirstName + " " + dataset.Users[index].LastName
		if strings.Contains(fullNameOfUser, query) || strings.Contains(dataset.Users[index].About, query) {
			counter += 1
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

func fillAllUsers(dataset Dataset, usersWeNeed []User) []User {
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

func SearchServer(w http.ResponseWriter, r *http.Request) {
	paramsQuery := r.URL.Query()

	// limit - максимальное количество записей, которое мы можем вернуть (+1, чтобы клиент у себя сортил флаг NextPage - true/false
	limit, err := strconv.Atoi(paramsQuery.Get("limit"))
	if err != nil {
		fmt.Fprintln(w, "Не получилось конвертировать параметр limit в число")
		return
	}
	if limit == 0 {
		fmt.Fprintln(w, "Лимит пользователей 0! Сервер не может вернуть ни одного пользователя!")
		return
	}
	offset := paramsQuery.Get("offset")           // позволяет получить список с определенного индекса в уже полученном отсортированном слайсе
	query := paramsQuery.Get("query")             // текст (подстрока), который нужно найти в нашем файле dataset.xml в полях Name/About
	order_field := paramsQuery.Get("order_field") // указывает по какому полю сортировать собранный слайс подходящих юзеров
	order_by := paramsQuery.Get("order_by")       // задает направление сортировки: 1 по возрастанию, 0 как встретилось, -1 по убыванию

	file, err := os.Open("dataset.xml") // err это переменная для ошибки
	if err != nil {
		fmt.Fprintln(w, "Сервер не смог прочитать файл")
	}
	defer file.Close()

	var dataset Dataset
	err = xml.NewDecoder(file).Decode(&dataset)
	if err != nil {
		fmt.Fprintln(w, "Сервер не смог распарсить файл")
	}

	// searcherParams.Add("limit", strconv.Itoa(req.Limit))
	//	searcherParams.Add("offset", strconv.Itoa(req.Offset))
	//	searcherParams.Add("query", req.Query)
	//	searcherParams.Add("order_field", req.OrderField)
	//	searcherParams.Add("order_by", strconv.Itoa(req.OrderBy))

}
