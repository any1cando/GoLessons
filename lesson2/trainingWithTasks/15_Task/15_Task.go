// 15. Обработка заказов с отменой — итоговая задача
// Доработай предыдущую программу:
// - Увеличь количество заказов до 100.
// - Обработка одного заказа занимает примерно 100 ms.
// - Ограничь время всей обработки контекстом на 350 ms.
// - При отмене прекрати выдавать новые заказы и прерви ожидание внутри работников.
// - Отправка заданий и результатов тоже должна учитывать отмену.
// - Собери результаты, которые успели поступить.
// - Дождись завершения всех горутин.
// - Выведи число обработанных заказов и причину остановки.
// Точное число результатов здесь может различаться. Главное — программа завершается, не зависает
// и не оставляет работников заблокированными.

package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type Order struct {
	ID     int
	Amount float64
}

func worker(ctx context.Context, orders <-chan Order, results chan<- Order, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		var orderInFunction Order

		// Ждём заказ или отмену.
		select {
		case <-ctx.Done():
			return
		case nextOrder, ok := <-orders:
			if !ok {
				return
			}
			orderInFunction = nextOrder
		}

		// Имитируем обработку, которую можно прервать.
		select {
		case <-ctx.Done():
			return
		case <-time.After(100 * time.Millisecond):
		}

		orderInFunction.Amount *= 0.9

		// Отправляем результат или завершаемся при отмене.
		select {
		case <-ctx.Done():
			return
		case results <- orderInFunction:
		}
	}
}

func main() {
	// Вставь сюда свой список из 100 заказов.
	orders := []Order{
		{1, 115},
		{2, 325},
		{3, 125},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 350*time.Millisecond)
	defer cancel()

	channelForOrders := make(chan Order)
	channelForResults := make(chan Order)

	var workersWG sync.WaitGroup
	var senderWG sync.WaitGroup

	// Три работника.
	for i := 0; i < 3; i++ {
		workersWG.Add(1)
		go worker(ctx, channelForOrders, channelForResults, &workersWG)
	}

	// Отправитель заказов тоже учитывает отмену.
	senderWG.Add(1)
	go func() {
		defer senderWG.Done()
		defer close(channelForOrders)

		for _, order := range orders {
			select {
			case <-ctx.Done():
				return
			case channelForOrders <- order:
			}
		}
	}()

	// Когда отправитель и работники завершились,
	// новых результатов больше не будет.
	go func() {
		senderWG.Wait()
		workersWG.Wait()
		close(channelForResults)
	}()

	processed := 0

	for result := range channelForResults {
		processed++
		fmt.Printf(
			"Заказ №%d: стоимость со скидкой — %.2f\n",
			result.ID,
			result.Amount,
		)
	}

	fmt.Println("Получено результатов:", processed)

	if err := ctx.Err(); err != nil {
		fmt.Println("Причина остановки:", err)
	} else {
		fmt.Println("Все заказы обработаны")
	}
}
