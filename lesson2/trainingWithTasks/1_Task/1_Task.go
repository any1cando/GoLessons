package main

import "fmt"

func totalAmountOnItems(pricePerItem float32, amount int, resultChannel chan<- float32) {
	resultChannel <- pricePerItem * float32(amount)
}

func main() {
	myChannel := make(chan float32)
	amountOfItems := 10
	var pricePerItem float32 = 100.00
	go totalAmountOnItems(pricePerItem, amountOfItems, myChannel)
	fmt.Println(int(<-myChannel))
}
