package tasks

import (
	"fmt"
)

type TaskStorage struct {
	storage []Task
	nextID  int
}

func NewStorage() *TaskStorage {
	return &TaskStorage{}
}

func (s *TaskStorage) AddTask(task Task) int {
	task.ID = s.nextID
	s.nextID++
	s.storage = append(s.storage, task)

	return task.ID
}

func (s *TaskStorage) DeleteTaskByID(id int) error {
	for index, task := range s.storage {
		if task.ID == id {
			s.storage = append(s.storage[:index], s.storage[index+1:]...)
			return nil
		}
	}

	return fmt.Errorf("task %d: ID does not exist", id)
}

func (s *TaskStorage) GetAllTasks() []Task {
	result := make([]Task, len(s.storage))
	copy(result, s.storage)
	return result
}
