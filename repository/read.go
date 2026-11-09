package repository

import (
	"awesomeProject/tasks"
	"context"
	"log"

	"github.com/jackc/pgx/v5"
)

func (s *PostgresStorage) GetAllTasks(ctx context.Context, done *bool) ([]tasks.Task, error) {
	var rows pgx.Rows
	var err error
	tasksList := make([]tasks.Task, 0)

	switch done {
	case nil:
		rows, err = s.pool.Query(ctx,
			`SELECT id, title, done
				 FROM tasks
				 ORDER BY id;`,
		)
		defer rows.Close()
	default:
		rows, err = s.pool.Query(
			ctx,
			`SELECT id, title, done
				 FROM tasks
				 WHERE done = $1
				 ORDER BY id;`,
			*done,
		)
		defer rows.Close()
	}

	if err != nil {
		log.Printf("query failed: %v", err)
		return nil, err
	}

	for rows.Next() {
		var task tasks.Task

		if err := rows.Scan(&task.ID, &task.Title, &task.Done); err != nil {
			log.Printf("cant scan row: %v", err)
			return nil, err
		}
		tasksList = append(tasksList, task)
	}

	return tasksList, nil
}
