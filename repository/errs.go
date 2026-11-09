package repository

import "fmt"

type ErrTaskNotFound struct {
	ID int
}

func (e *ErrTaskNotFound) Error() string {
	return fmt.Sprintf("task %d ID not found", e.ID)
}
