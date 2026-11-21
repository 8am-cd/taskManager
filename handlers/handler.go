package handlers

import (
	"awesomeProject/tasks"
	"awesomeProject/users"
	"time"
)

type Storage interface {
	GetAllTasks(userID int, done *bool) ([]tasks.Task, error)
	AddTask(userID int, task tasks.Task) (int, error)
	PatchTask(userID, id int, title *string, done *bool) error
	DeleteTaskByID(userID, id int) error

	AddUser(user *users.User) error
	LoginUser(user *users.User) (string, time.Time, error)
	LogoutUser(token string) error
	SessionActive(tokenHash string) (int, error)
}

type Handler struct {
	storage Storage
}

func NewHandler(storage Storage) Handler {
	return Handler{storage: storage}
}
