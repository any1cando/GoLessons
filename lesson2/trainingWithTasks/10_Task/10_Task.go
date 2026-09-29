// 10. Остановка фоновой работы — context.WithCancel
//Фоновый процесс регулярно проверяет новые сообщения, пока пользователь не выйдет.
//- Создай отменяемый контекст.
//- Запусти горутину с тикером на 100 ms.
//- На каждом тике она выводит «Проверяю сообщения».
//- В main примерно через 350 ms вызови отмену.
//- Горутина должна получить сигнал через ctx.Done(), остановить тикер и завершиться.
//- Дождись её завершения и выведи «Приложение закрыто».

package main

import (
	context2 "context"
	"fmt"
	"sync"
	"time"
)

func tickerSender(ctx context2.Context, ticker *time.Ticker, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case <-ticker.C:
			fmt.Println("Проверяю сообщения")
		case <-ctx.Done():
			ticker.Stop()
			return
		}
	}
}

func main() {
	context, cancel := context2.WithCancel(context2.Background())
	ticker := time.NewTicker(time.Millisecond * 100)
	waitGroup := &sync.WaitGroup{}

	waitGroup.Add(1)
	go tickerSender(context, ticker, waitGroup)

	time.Sleep(time.Millisecond * 300)
	cancel()

	waitGroup.Wait()
	fmt.Println("Приложение закрыто")
}
