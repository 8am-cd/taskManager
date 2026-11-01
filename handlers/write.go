package handlers

import (
	"awesomeProject/tasks"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func (h *Handler) PostTask(w http.ResponseWriter, req *http.Request) {
	var input struct {
		Title string `json:"title"`
	}

	if err := json.NewDecoder(req.Body).Decode(&input); err != nil {
		http.Error(w, "can't read json", http.StatusBadRequest)
		return
	}

	input.Title = strings.TrimSpace(input.Title)
	if len(input.Title) == 0 {
		http.Error(w, "title is empty", http.StatusBadRequest)
		return
	}

	id := h.TaskStorage.AddTask(tasks.NewTask(input.Title))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(map[string]int{"id": id}); err != nil {
		log.Printf("output ID task: %v", err)
	}
}

func (h *Handler) DeleteTask(w http.ResponseWriter, req *http.Request) {
	id, err := strconv.Atoi(req.PathValue("id"))
	if err != nil || id < 0 {
		http.Error(w, "id is incorrect", http.StatusBadRequest)
		return
	}

	if err := h.TaskStorage.DeleteTaskByID(id); err != nil {
		http.Error(w, "task does not exist", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
