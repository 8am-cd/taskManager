package handlers

import (
	"awesomeProject/repository"
)

type Handler struct {
	postgresStorage *repository.PostgresStorage
}

func NewHandler(storage *repository.PostgresStorage) Handler {
	return Handler{postgresStorage: storage}
}
