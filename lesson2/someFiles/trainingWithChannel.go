package main

import "fmt"

func main() {

	var in chan int = make(chan int, 1) // убрали буфер на 5

	go func(channelOnlyForWrite chan<- int) {
		for i := 0; i <= 10; i++ {
			fmt.Println("before", i)
			channelOnlyForWrite <- i
			fmt.Println("after", i)
		}
		close(channelOnlyForWrite)
		fmt.Println("Generator finished!") // Данная фраза не всегда выводится, потому что горутина мейна не
		// ждет дозавершения этой горутины. Когда проверка на isOpen не пройдет, и мейн выйдет из цикла - он может
		// просто успеть завершить программу, а до выполнения Generator finished может просто не дойти

	}(in)

	for {
		numberFromChannel, isOpen := <-in

		if !isOpen {
			break
		}

		fmt.Println("\tget", numberFromChannel)
	}

}
