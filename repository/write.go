package repository

import (
	"awesomeProject/tasks"
	"context"
	"errors"
	"log"
)

func (s *PostgresStorage) AddTask(ctx context.Context, task tasks.Task) (int, error) {
	var id int

	err := s.pool.QueryRow(
		ctx,
		`INSERT INTO tasks (title, done) VALUES ($1, false) RETURNING id;`,
		task.Title,
	).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (s *PostgresStorage) PatchTask(ctx context.Context, id int, title *string, done *bool) error {
	if title != nil && len(*title) == 0 {
		title = nil
	}

	tag, err := s.pool.Exec(
		ctx,
		`UPDATE tasks
			 SET title = COALESCE($1::text, title),
			     done = COALESCE($2::boolean, done)
			 WHERE id = $3`,
		title,
		done,
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

func (s *PostgresStorage) DeleteTaskByID(ctx context.Context, id int) error {
	tag, err := s.pool.Exec(
		ctx,
		`DELETE FROM tasks WHERE id = $1`,
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
