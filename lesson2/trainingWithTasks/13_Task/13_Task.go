//13. Продажа билетов — sync.Mutex
//Есть 20 билетов и 50 покупателей, каждый хочет купить один билет.
//- Каждый покупатель — отдельная горутина.
//- Покупатель проверяет наличие билетов и, если они есть, уменьшает остаток на один.
//- Защити общие данные мьютексом.
//- Посчитай число успешных покупок.
//- Дождись всех покупателей и выведи итог.
//Должно быть: 20 успешных покупок и 0 оставшихся билетов. Проверка наличия и уменьшение остатка должны выполняться
//как одна защищённая операция. Проверь через -race.

package main

import (
	"fmt"
	"sync"
)

func main() {

	ticketsForSale := 20
	ticketsWhichWereBought := 0
	mtxLock := &sync.Mutex{}
	wg := &sync.WaitGroup{}

	for range 50 {

		wg.Add(1)
		go func() {
			defer wg.Done()
			mtxLock.Lock()
			if ticketsForSale == 0 {
				mtxLock.Unlock()
				return
			}
			ticketsWhichWereBought++
			ticketsForSale--
			mtxLock.Unlock()
		}()
	}
	wg.Wait()
	fmt.Println("Дождались всех покупателей. Билетов осталось:", ticketsForSale)
	fmt.Println("Билетов продано:", ticketsWhichWereBought)
}
