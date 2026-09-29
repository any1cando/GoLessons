// 3. Сбор показаний — закрытие канала
//Датчик передаёт последовательность измерений.
//- В отдельной горутине отправь в канал значения 18, 21, 19, 24, 20.
//- После отправки всех значений закрой канал со стороны отправителя.
//- В main прочитай значения через range.
//- Выведи количество измерений и среднее значение.
//Ожидается: 5 измерений, среднее — 20.4.

package main

import "fmt"

func sender(ch chan<- int) {
	// 18 21 19 24 20
	ch <- 18
	ch <- 21
	ch <- 19
	ch <- 24
	ch <- 20
	close(ch)
}

func main() {
	myChannel := make(chan int)
	go sender(myChannel)
	allNumbers := 0.00
	counter := 0.00
	for number := range myChannel {
		allNumbers += float64(number)
		counter++
	}
	fmt.Printf("Количество - %v, среднее - %v", counter, allNumbers/counter)
}
