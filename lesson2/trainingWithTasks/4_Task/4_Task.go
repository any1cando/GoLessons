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
