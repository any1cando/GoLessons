// 9. Отложенное действие — time.AfterFunc
//После оформления заказа нужно отправить уведомление с небольшой задержкой.
//- Запланируй функцию через time.AfterFunc на 200 ms.
//- Внутри неё выведи «Уведомление отправлено».
//- Дождись завершения функции через канал или WaitGroup.
//- Затем выведи «Готово».
//Дополнение: сделай отдельный вариант, где запланированное действие отменяется сразу после создания таймера.
//Программа должна корректно завершаться и при отмене.

package main

import (
	"fmt"
	"sync"
	"time"
)

func sendNotification(wg *sync.WaitGroup) {
	defer wg.Done()
	fmt.Println("Уведомление отправлено")
}

func main() {
	waitGroup := &sync.WaitGroup{}
	waitGroup.Add(1) // Ждём одно уведомление

	time.AfterFunc(200*time.Millisecond, func() {
		sendNotification(waitGroup)
	})

	waitGroup.Wait()
	fmt.Println("Готово")
}
