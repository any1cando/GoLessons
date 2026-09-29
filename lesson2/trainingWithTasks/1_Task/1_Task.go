// 1. Расчёт стоимости — горутина и канал
//Магазину нужно посчитать стоимость покупки в отдельной горутине.
//- Создай функцию, принимающую цену, количество товара и канал результата.
//- Запусти её как горутину.
//- Передай через канал общую стоимость.
//- В main получи результат и выведи его.
//Проверь на цене 150 и количестве 3: результат должен быть 450. Канал сделай небуферизированным.

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
