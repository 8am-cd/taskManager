package handlers

import (
	"awesomeProject/tasks"
	"encoding/json"
	"log"
	"net/http"
)

func (h *Handler) GetAllTasks(w http.ResponseWriter, req *http.Request) {
	var list struct {
		Tasks []tasks.Task `json:"tasks"`
	}
	list.Tasks = h.TaskStorage.GetAllTasks()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(list); err != nil {
		log.Printf("output of task storage: %v", err)
	}
}
