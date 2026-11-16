package repository

import (
	"awesomeProject/tasks"
	"context"
	"errors"
	"log"
)

func (s *PostgresStorage) AddTask(userID int, task tasks.Task) (int, error) {
	var id int

	err := s.pool.QueryRow(
		context.Background(),
		`INSERT INTO tasks (title, done, user_id) VALUES ($1, false, $2) RETURNING id;`,
		task.Title,
		userID,
	).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (s *PostgresStorage) PatchTask(userID int, id int, title *string, done *bool) error {
	if title != nil && len(*title) == 0 {
		title = nil
	}

	tag, err := s.pool.Exec(
		context.Background(),
		`UPDATE tasks
			 SET title = COALESCE($1::text, title),
			     done = COALESCE($2::boolean, done)
			 WHERE user_id = $3 AND id = $4`,
		title,
		done,
		userID,
		id,
	)

	if err != nil {
		log.Printf("failed patch task query: %v", err)
		return err
	}

	if tag.RowsAffected() == 0 {
		log.Println("id not found")
		return errors.New("not found")
	}

	return nil
}

func (s *PostgresStorage) DeleteTaskByID(userID int, id int) error {
	tag, err := s.pool.Exec(
		context.Background(),
		`DELETE FROM tasks WHERE user_id = $1 AND id = $2`,
		userID,
		id,
	)
	if err != nil {
		log.Printf("failed delete task by id query: %v", err)
		return err
	}

	if tag.RowsAffected() == 0 {
		log.Println("id not found")
		return &ErrTaskNotFound{id}
	}

	return nil
}
