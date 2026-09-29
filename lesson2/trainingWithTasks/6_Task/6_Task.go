// 6. Ограничение ожидания — time.After
//Пользователь не должен ждать ответ сервиса дольше 200 ms.
//- Запусти горутину, которая имитирует запрос и отправляет строку с результатом.
//- В main через select ожидай либо результат, либо истечение 200 ms.
//- Выведи ответ или «Время ожидания истекло».
//- Проверь запросы длительностью 50 ms и 500 ms.
//- Отправка позднего ответа не должна навсегда блокировать горутину.

package main

import (
	"fmt"
	"time"
)

func sendRequest(ch chan string) {
	time.Sleep(time.Millisecond * 500)
	ch <- "OK"
}

func main() {
	timer := time.After(time.Millisecond * 200) // будем ждать только 200 мс
	myChannel := make(chan string, 1)
	go sendRequest(myChannel)
	select {
	case <-timer:
		fmt.Println("Время ожидания истекло")
	case answer := <-myChannel:
		fmt.Printf("Answer from server is %s", answer)
	}

}
