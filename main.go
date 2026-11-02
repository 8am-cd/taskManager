package main

import (
	"awesomeProject/handlers"
	"awesomeProject/tasks"
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	taskStorage := tasks.NewStorage()
	handler := handlers.NewHandler(taskStorage)

	mux.HandleFunc("GET /tasks", handler.GetAllTasks)
	mux.HandleFunc("POST /tasks", handler.PostTask)
	mux.HandleFunc("DELETE /tasks/{id}", handler.DeleteTask)

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	log.Println("Сервер слушает порт 8080")
	log.Fatal(server.ListenAndServe())
}
