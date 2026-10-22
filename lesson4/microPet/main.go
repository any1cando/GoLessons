package main

import (
	"firstGoProject/lesson4/microPet/handler"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
)

func main() {
	var tasks = make(map[int]*handler.Task)
	mu := &sync.Mutex{}
	router := mux.NewRouter().StrictSlash(true)

	myHandler := handler.Handler{
		Tasks:  tasks,
		Mu:     mu,
		Router: router,
	}

	router.HandleFunc("/tasks/", myHandler.GetAllTasks).Methods("GET")
	router.HandleFunc("/tasks/{id}", myHandler.GetCertainTask).Methods("GET")
	router.HandleFunc("/tasks/", myHandler.CreateTask).Methods("POST")
	router.HandleFunc("/tasks/{id}", myHandler.ChangeTask).Methods("PUT")
	router.HandleFunc("tasks/{id}", myHandler.DeleteTask).Methods("DELETE")

	myServer := http.Server{
		Addr:         ":8080",
		Handler:      &myHandler,
		ReadTimeout:  time.Second * 10,
		WriteTimeout: time.Second * 10,
	}

	fmt.Println("Starting server at port 8080...")
	err := myServer.ListenAndServe()
	if err != nil {
		fmt.Println("Can't start server. The error is", err)
		return
	}
}
