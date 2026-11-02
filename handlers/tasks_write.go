package handlers

import (
	"awesomeProject/tasks"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
)

func (h *Handler) PostTask(w http.ResponseWriter, req *http.Request) {
	req.Body = http.MaxBytesReader(w, req.Body, 1<<10)

	var input struct {
		Title string `json:"title"`
	}

	dec := json.NewDecoder(req.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&input); err != nil {
		if _, ok := errors.AsType[*http.MaxBytesError](err); ok {
			WriteError(w, "title size is too large", http.StatusRequestEntityTooLarge)
			return
		}

		WriteError(w, "can't read json", http.StatusBadRequest)
		return
	}

	var extra any
	if err := dec.Decode(&extra); err != io.EOF {
		WriteError(w, "allowed only one field title", http.StatusBadRequest)
		return
	}

	input.Title = strings.TrimSpace(input.Title)
	if len(input.Title) == 0 {
		WriteError(w, "title is empty", http.StatusBadRequest)
		return
	}

	id := h.taskStorage.AddTask(tasks.NewTask(input.Title))

	WriteJSON(w, http.StatusCreated, map[string]int{"id": id})
}

func (h *Handler) DeleteTask(w http.ResponseWriter, req *http.Request) {
	id, err := strconv.Atoi(req.PathValue("id"))
	if err != nil || id < 0 {
		WriteError(w, "id is incorrect", http.StatusBadRequest)
		return
	}

	if err := h.taskStorage.DeleteTaskByID(id); err != nil {
		WriteError(w, "task does not exist", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
