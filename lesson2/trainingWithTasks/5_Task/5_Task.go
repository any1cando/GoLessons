package main

import (
	"fmt"
	"sync"
	"time"
)

func askFirstService(wg *sync.WaitGroup, channel chan<- int) {
	defer wg.Done()
	time.Sleep(time.Millisecond * 100)
	channel <- 100
	close(channel)
}

func askSecondService(wg *sync.WaitGroup, channel chan<- int) {
	defer wg.Done()
	time.Sleep(time.Millisecond * 300)
	channel <- 95
	close(channel)
}

func main() {
	ch1 := make(chan int, 1)
	ch2 := make(chan int, 1)
	waitGroup := &sync.WaitGroup{}
	waitGroup.Add(1)
	waitGroup.Add(1)
	go askFirstService(waitGroup, ch1)
	go askSecondService(waitGroup, ch2)
	ch1Closed, ch2Closed := false, false
	for !ch1Closed || !ch2Closed {
		select {
		case answer1, ok := <-ch1:
			if !ok {
				ch1Closed = true
			} else {
				fmt.Printf("Источник - %v, цена - %d\n", "Сервис №1", answer1)
			}
		case answer2, ok := <-ch2:
			if !ok {
				ch2Closed = true
			} else {
				fmt.Printf("Источник - %v, цена - %d\n", "Сервис №2", answer2)
			}
		}
	}
	waitGroup.Wait()
	fmt.Println("Мы дождались обоих серверов!")
}

// Если мы создаем канал НЕбуферизированный, то в таком случае мы можем передавать инфу только из рук в руки, то есть
// на этом канале нет коробочки, куда можно было бы положить что-то и идти дальше по своим делам
