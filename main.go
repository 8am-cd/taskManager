package main

import (
	"awesomeProject/handlers"
	"awesomeProject/repository"
	"context"
	"log"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	pool, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("unable to create new pool: %v\n", err)
	}
	defer pool.Close()

	if err := pool.Ping(context.Background()); err != nil {
		log.Fatalf("database is unavailable: %v\n", err)
	}

	mux := http.NewServeMux()
	registerFrontend(mux)
	storage := repository.NewStorage(pool)
	handler := handlers.NewHandler(storage)

	mux.HandleFunc("GET /tasks", handler.GetAllTasks)
	mux.HandleFunc("POST /tasks", handler.PostTask)
	mux.HandleFunc("PATCH /tasks/{id}", handler.PatchTask)
	mux.HandleFunc("DELETE /tasks/{id}", handler.DeleteTaskByID)

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	log.Println("Сервер слушает порт 8080")
	log.Fatal(server.ListenAndServe())
}
