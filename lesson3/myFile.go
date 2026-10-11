package main

import (
	"fmt"
	"net/http"
	"time"
)

func handler123(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Страница главная!")
}

func handler133(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Страница пользователя!")
}

func main() {

	myMux := http.NewServeMux()
	myMux.HandleFunc("/", handler123)
	myMux.HandleFunc("/user/", handler133)

	server := http.Server{
		Addr:         ":8080",
		Handler:      myMux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	server.ListenAndServe()

}
