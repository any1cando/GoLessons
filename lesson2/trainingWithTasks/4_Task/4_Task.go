// 4. Подготовка файлов — sync.WaitGroup
//Перед запуском приложения нужно подготовить три файла.
//- Создай функцию подготовки файла, принимающую его название.
//- Запусти три горутины для users.csv, orders.csv и products.csv.
//- Имитируй работу задержками разной длины.
//- Каждая горутина должна вывести сообщение о завершении.
//- Через WaitGroup дождись всех трёх.
//- Только после этого выведи «Все файлы подготовлены».
//Реально создавать файлы не нужно.

package main

import (
	"fmt"
	"sync"
	"time"
)

func fileReader(strNameOfFile string, waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()
	switch strNameOfFile {
	case "users.csv":
		time.Sleep(time.Second * 2)
	case "orders.csv":
		time.Sleep(time.Second * 4)
	case "products.csv":
		time.Sleep(time.Second * 5)
	}
	time.Sleep(time.Second * 2)
	fmt.Printf("Файл %v прочитан, задача снята\n", strNameOfFile)
}

func main() {
	myWaitGroup := &sync.WaitGroup{}
	listOfFiles := []string{"users.csv", "orders.csv", "products.csv"}
	for _, file := range listOfFiles {
		myWaitGroup.Add(1)
		go fileReader(file, myWaitGroup)
	}
	myWaitGroup.Wait()
	fmt.Println("Все файлы подготовлены!")
}
