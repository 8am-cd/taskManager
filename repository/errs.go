package repository

import "fmt"

type ErrTaskNotFound struct {
	ID int
}

func (e *ErrTaskNotFound) Error() string {
	return fmt.Sprintf("task %d ID not found", e.ID)
}

type ErrUserExists struct {
	Email string
}

func (e *ErrUserExists) Error() string {
	return fmt.Sprintf("user with email %s already exists", e.Email)
}

type ErrUserNotExists struct {
	Email string
}

func (e *ErrUserNotExists) Error() string {
	return fmt.Sprintf("user with email %s not exists", e.Email)
}

type ErrWrongPassword struct {
}

func (e *ErrWrongPassword) Error() string {
	return fmt.Sprintf("wrong password")
}
