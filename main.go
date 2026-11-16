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

	mux.Handle("POST /auth/login", http.HandlerFunc(handler.LoginUser))
	mux.Handle("POST /auth/register", http.HandlerFunc(handler.RegisterUser))
	mux.Handle("POST /auth/logout", http.HandlerFunc(handler.LogoutUser))
	mux.Handle("GET /tasks", handler.Authorization(http.HandlerFunc(handler.GetAllTasks)))
	mux.Handle("POST /tasks", handler.Authorization(http.HandlerFunc(handler.PostTask)))
	mux.Handle("PATCH /tasks/{id}", handler.Authorization(http.HandlerFunc(handler.PatchTask)))
	mux.Handle("DELETE /tasks/{id}", handler.Authorization(http.HandlerFunc(handler.DeleteTaskByID)))

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	log.Println("Сервер слушает порт 8080")
	log.Fatal(server.ListenAndServe())
}
