//14. Обработка заказов — пул работников
//Есть десять заказов, но одновременно обрабатывать можно только три.
//- Создай структуру Order с полями ID и Amount.
//- Подготовь десять заказов.
//- Создай канал заданий и канал результатов.
//- Запусти ровно три горутины-работника.
//- Каждый работник получает заказы из общего канала и рассчитывает стоимость со скидкой 10%.
//- Результат должен содержать ID заказа и итоговую стоимость.
//- В main собери и выведи все результаты.
//- Канал результатов закрой только после завершения всех работников.
//Каждый заказ должен быть обработан ровно один раз. Порядок результатов может отличаться от порядка заказов.

package main

import (
	"fmt"
	"sync"
)

type Order struct {
	ID     int
	amount float64
}

func workForWorker(chanRead <-chan Order, chanWrite chan<- Order, wg *sync.WaitGroup) {
	defer wg.Done()
	for order := range chanRead {
		order.amount *= 0.9
		chanWrite <- order
	}
}

func printResult(orderToPrint Order) {
	fmt.Printf("Заказ с номером %d: результат - %f\n", orderToPrint.ID, orderToPrint.amount)
}

func main() {
	channelForOrders := make(chan Order)
	channelForResults := make(chan Order)
	waitGroup := &sync.WaitGroup{}

	order1 := Order{1, 115}
	order2 := Order{2, 325}
	order3 := Order{3, 125}
	order4 := Order{4, 545}
	order5 := Order{5, 875}
	order6 := Order{6, 576}
	order7 := Order{7, 512}
	order8 := Order{8, 532}
	order9 := Order{9, 55}
	order10 := Order{10, 50}
	listForOrders := []Order{order1, order2, order3, order4, order5, order6, order7, order8, order9, order10}

	for i := 1; i <= 3; i++ {
		waitGroup.Add(1)
		go workForWorker(channelForOrders, channelForResults, waitGroup)
	}

	go func() {
		for _, order := range listForOrders {
			channelForOrders <- order
		}
		close(channelForOrders)
	}() // хороший паттерн, который заключается в том, что мы передаем инфу работникам и под конец этой горутины
	// закрываем канал, чтобы не было deadlock

	go func() {
		waitGroup.Wait()
		close(channelForResults)
	}()

	for resultString := range channelForResults {
		printResult(resultString)
	}

	fmt.Println("Программа завершена")
}
