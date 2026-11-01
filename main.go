package main

import (
	"awesomeProject/handlers"
	"awesomeProject/tasks"
	"log"
	"net/http"
)

func main() {
	taskStorage := tasks.NewStorage()
	handler := handlers.NewHandler(taskStorage)

	http.HandleFunc("GET /tasks", handler.GetAllTasks)
	http.HandleFunc("POST /tasks", handler.PostTask)
	http.HandleFunc("DELETE /tasks/{id}", handler.DeleteTask)

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}
