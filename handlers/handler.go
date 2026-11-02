package handlers

import "awesomeProject/tasks"

type Handler struct {
	taskStorage *tasks.TaskStorage
}

func NewHandler(storage *tasks.TaskStorage) Handler {
	return Handler{taskStorage: storage}
}
