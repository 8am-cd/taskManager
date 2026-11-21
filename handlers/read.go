package handlers

import (
	"awesomeProject/tasks"
	"net/http"
)

func (h *Handler) GetAllTasks(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value(keyUserID).(int)

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

	tasksList, err := h.storage.GetAllTasks(userID, donePtr)
	if err != nil {
		WriteError(w, "cant get tasks", http.StatusInternalServerError)
		return
	}

	WriteJSON(w, http.StatusOK, struct {
		Tasks []tasks.Task `json:"tasks"`
	}{
		Tasks: tasksList,
	})
}
