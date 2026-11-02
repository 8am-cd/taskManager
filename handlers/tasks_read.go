package handlers

import (
	"awesomeProject/tasks"
	"net/http"
)

func (h *Handler) GetAllTasks(w http.ResponseWriter, req *http.Request) {
	WriteJSON(w, http.StatusOK, struct {
		Tasks []tasks.Task `json:"tasks"`
	}{
		Tasks: h.taskStorage.GetAllTasks(),
	})
}
