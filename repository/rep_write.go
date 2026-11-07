package repository

import (
	"awesomeProject/tasks"
	"context"
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

func (s *PostgresStorage) PatchTask(ctx context.Context, id int, title string, done bool) error {
	if len(title) != 0 {
		_, err := s.pool.Exec(
			ctx,
			`UPDATE tasks SET title = $1 WHERE id = $2`,
			title,
			id,
		)
		if err != nil {
			log.Printf("failed patch title task query: %v", err)
			return err
		}
	}

	if done == true {
		_, err := s.pool.Exec(
			ctx,
			`UPDATE tasks SET done = $1 WHERE id = $2`,
			done,
			id,
		)
		if err != nil {
			log.Printf("failed patch done flag task query: %v", err)
			return err
		}
	}

	return nil
}

func (s *PostgresStorage) DeleteTaskByID(ctx context.Context, id int) error {
	_, err := s.pool.Exec(
		ctx,
		`DELETE FROM tasks WHERE id = $1`,
		id,
	)
	if err != nil {
		log.Printf("failed delete task by id query: %v", err)
		return err
	}

	return nil
}
