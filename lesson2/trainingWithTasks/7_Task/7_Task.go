// 7. Напоминание с отменой — time.Timer
//Пользователю показывают напоминание, если он долго не подтверждает действие.
//- Создай таймер на 300 ms.
//- Отдельная горутина имитирует подтверждение пользователя через канал.
//- Если подтверждение пришло раньше — останови таймер и выведи «Подтверждено».
//- Если сработал таймер — выведи «Напоминание: подтвердите действие».
//- Проверь подтверждение через 100 ms и через 500 ms.
//Достаточно обработать первое событие. Поздняя отправка подтверждения не должна блокировать горутину.

package main

import (
	"fmt"
	"time"
)

func askUser(channelForAnswer chan string) {
	time.Sleep(time.Millisecond * 50)
	channelForAnswer <- "CONFIRMED"
}

func main() {
	channel := make(chan string, 1)
	timer := time.NewTimer(time.Millisecond * 300)
	go askUser(channel)

	select {
	case <-timer.C:
		fmt.Println("Подтвердите действие!")
	case answer := <-channel:
		timer.Stop()
		fmt.Printf("Ответ пользователя - %s\n", answer)
	}
	fmt.Println("Main завершил свою работу")
}
