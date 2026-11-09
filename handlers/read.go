package handlers

import (
	"awesomeProject/tasks"
	"context"
	"net/http"
)

func (h *Handler) GetAllTasks(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	doneStr := params.Get("done")
	var done bool
	var donePtr *bool

	switch doneStr {
	case "true":
		done = true
		donePtr = &done
	case "false":
		done = false
		donePtr = &done
	}

	tasksList, err := h.postgresStorage.GetAllTasks(context.Background(), donePtr)
	if err != nil{
		WriteError(w, "cant get tasks", http.StatusInternalServerError)
		return
	}

	WriteJSON(w, http.StatusOK, struct {
		Tasks []tasks.Task `json:"tasks"`
	}{
		Tasks: tasksList,
	})
}
