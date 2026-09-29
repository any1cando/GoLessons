package main

import (
	"fmt"
	"time"
)

func sendRequest(ch chan string) {
	time.Sleep(time.Millisecond * 500)
	ch <- "OK"
}

func main() {
	timer := time.After(time.Millisecond * 200) // будем ждать только 200 мс
	myChannel := make(chan string, 1)
	go sendRequest(myChannel)
	select {
	case <-timer:
		fmt.Println("Время ожидания истекло")
	case answer := <-myChannel:
		fmt.Printf("Answer from server is %s", answer)
	}

}
