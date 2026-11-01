package handlers

import "awesomeProject/tasks"

type Handler struct {
	TaskStorage *tasks.TaskStorage
}

func NewHandler(storage *tasks.TaskStorage) Handler {
	return Handler{TaskStorage: storage}
}
