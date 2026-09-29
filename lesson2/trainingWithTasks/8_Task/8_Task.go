// 8. Проверка состояния — time.Ticker
//Нужно периодически проверять состояние сервиса.
//- Создай тикер с интервалом 100 ms.
//- На каждом тике выводи «Проверка №N».
//- После пятой проверки останови тикер и заверши программу.
//- Не используй Sleep для организации периодичности.

package main

import (
	"fmt"
	"time"
)

func main() {
	ticker := time.NewTicker(time.Millisecond * 100)
	for i := 0; i <= 10; i++ {
		if i == 5 {
			ticker.Stop()
			break
		}
		fmt.Printf("Проверка №%d: %v\n", i+1, <-ticker.C)
	}
	fmt.Println("Main завершил свою работу")
}
