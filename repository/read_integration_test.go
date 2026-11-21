package repository_test

import (
	"context"
	"crypto/rand"
	"os"
	"slices"
	"testing"
	"time"

	"awesomeProject/repository"
	"awesomeProject/tasks"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestGetAllTasksIntegration(t *testing.T) {
	config, err := pgxpool.ParseConfig(os.Getenv("TEST_DATABASE_URL"))
	if err != nil {
		t.Fatal("invalid TEST_DATABASE_URL")
	}
	schema := "taskmanager_test_" + rand.Text()
	config.ConnConfig.RuntimeParams["search_path"] = pgx.Identifier{schema}.Sanitize()
	config.ConnConfig.RuntimeParams["statement_timeout"] = "5000"

	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		t.Fatalf("create test pool: %v", err)
	}
	t.Cleanup(pool.Close)

	schemaName := pgx.Identifier{schema}.Sanitize()
	if _, err := pool.Exec(ctx, "CREATE SCHEMA "+schemaName); err != nil {
		t.Fatalf("create test schema: %v", err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if _, err := pool.Exec(ctx, "DROP SCHEMA "+schemaName+" CASCADE"); err != nil {
			t.Errorf("drop test schema: %v", err)
		}
	})

	schemaSQL, err := os.ReadFile("../init.sql")
	if err != nil {
		t.Fatalf("read init.sql: %v", err)
	}
	if _, err := pool.Exec(ctx, string(schemaSQL)); err != nil {
		t.Fatalf("apply init.sql: %v", err)
	}

	const aliceID, bobID = 1, 2
	aliceOpen := tasks.Task{ID: 10, Title: "Read Go docs"}
	aliceDone := tasks.Task{ID: 30, Title: "Write unit tests", Done: true}
	bobDone := tasks.Task{ID: 20, Title: "Set up PostgreSQL", Done: true}

	if _, err := pool.Exec(ctx, `INSERT INTO users (id, email)
		VALUES ($1, 'alice@example.com'), ($2, 'bob@example.com')`, aliceID, bobID); err != nil {
		t.Fatalf("seed users: %v", err)
	}
	if _, err := pool.Exec(ctx, `INSERT INTO tasks (id, title, done, user_id)
		VALUES ($1, $2, $3, $4), ($5, $6, $7, $4), ($8, $9, $10, $11)`,
		aliceDone.ID, aliceDone.Title, aliceDone.Done, aliceID,
		aliceOpen.ID, aliceOpen.Title, aliceOpen.Done,
		bobDone.ID, bobDone.Title, bobDone.Done, bobID,
	); err != nil {
		t.Fatalf("seed tasks: %v", err)
	}

	storage := repository.NewStorage(pool)
	done, open := true, false
	cases := []struct {
		name   string
		userID int
		done   *bool
		want   []tasks.Task
	}{
		{"all tasks ordered by ID", aliceID, nil, []tasks.Task{aliceOpen, aliceDone}},
		{"completed tasks of the current user", aliceID, &done, []tasks.Task{aliceDone}},
		{"open tasks of the current user", aliceID, &open, []tasks.Task{aliceOpen}},
		{"second user's tasks", bobID, nil, []tasks.Task{bobDone}},
		{"no matching tasks", bobID, &open, []tasks.Task{}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := storage.GetAllTasks(tc.userID, tc.done)
			if err != nil {
				t.Fatalf("GetAllTasks: %v", err)
			}
			if got == nil || !slices.Equal(got, tc.want) {
				t.Errorf("GetAllTasks = %#v, want %#v", got, tc.want)
			}
		})
	}
}
