package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"awesomeProject/tasks"
)

type taskListStorageStub struct {
	Storage // В этом тесте вызывается только GetAllTasks.
	list    func(userID int, done *bool) ([]tasks.Task, error)
}

func (s taskListStorageStub) GetAllTasks(userID int, done *bool) ([]tasks.Task, error) {
	return s.list(userID, done)
}

func TestGetAllTasks(t *testing.T) {
	const userID = 42
	done, undone := true, false

	cases := []struct {
		name       string
		path       string
		wantDone   *bool
		storeTasks []tasks.Task
		storeErr   error
		wantStatus int
		wantJSON   string
	}{
		{
			name:       "list without filter",
			path:       "/tasks",
			storeTasks: []tasks.Task{{ID: 7, Title: "Изучить Go", Done: true}},
			wantStatus: http.StatusOK,
			wantJSON:   `{"tasks":[{"id":7,"title":"Изучить Go","done":true}]}`,
		},
		{
			name:       "filter by completed",
			path:       "/tasks?done=true",
			wantDone:   &done,
			storeTasks: []tasks.Task{{ID: 7, Title: "Изучить Go", Done: true}},
			wantStatus: http.StatusOK,
			wantJSON:   `{"tasks":[{"id":7,"title":"Изучить Go","done":true}]}`,
		},
		{
			name:       "filter by incomplete",
			path:       "/tasks?done=false",
			wantDone:   &undone,
			storeTasks: []tasks.Task{{ID: 8, Title: "Написать тест", Done: false}},
			wantStatus: http.StatusOK,
			wantJSON:   `{"tasks":[{"id":8,"title":"Написать тест","done":false}]}`,
		},
		{
			name:       "empty list",
			path:       "/tasks",
			storeTasks: []tasks.Task{},
			wantStatus: http.StatusOK,
			wantJSON:   `{"tasks":[]}`,
		},
		{
			name:       "storage failure",
			path:       "/tasks",
			storeErr:   errors.New("database unavailable"),
			wantStatus: http.StatusInternalServerError,
			wantJSON:   `{"error":"cant get tasks"}`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			h := NewHandler(taskListStorageStub{
				list: func(gotUserID int, gotDone *bool) ([]tasks.Task, error) {
					calls++
					if gotUserID != userID {
						t.Errorf("user ID = %d, want %d", gotUserID, userID)
					}
					if !reflect.DeepEqual(gotDone, tc.wantDone) {
						t.Errorf("done filter = %v, want %v", formatDone(gotDone), formatDone(tc.wantDone))
					}
					return tc.storeTasks, tc.storeErr
				},
			})
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			req = req.WithContext(context.WithValue(req.Context(), keyUserID, userID))
			res := httptest.NewRecorder()

			h.GetAllTasks(res, req)

			if calls != 1 {
				t.Errorf("storage calls = %d, want 1", calls)
			}
			if res.Code != tc.wantStatus {
				t.Fatalf("status = %d, want %d; body: %s", res.Code, tc.wantStatus, res.Body)
			}
			if got := res.Result().Header.Get("Content-Type"); got != "application/json" {
				t.Errorf("Content-Type = %q, want application/json", got)
			}

			var got, want any
			if err := json.Unmarshal(res.Body.Bytes(), &got); err != nil {
				t.Fatalf("invalid response JSON: %v; body: %s", err, res.Body)
			}
			if err := json.Unmarshal([]byte(tc.wantJSON), &want); err != nil {
				t.Fatalf("invalid expected JSON: %v", err)
			}
			if !reflect.DeepEqual(got, want) {
				t.Errorf("body = %s, want %s", res.Body, tc.wantJSON)
			}
		})
	}
}

func formatDone(done *bool) any {
	if done == nil {
		return nil
	}
	return *done
}
