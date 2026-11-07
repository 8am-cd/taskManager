package handlers

import (
	"awesomeProject/tasks"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
)

func (h *Handler) PostTask(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<10)

	var input struct {
		Title string `json:"title"`
	}

	dec := json.NewDecoder(r.Body)
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

	id, err := h.postgresStorage.AddTask(context.Background(), tasks.NewTask(input.Title))
	if err != nil {
		WriteError(w, "cant add task", http.StatusInternalServerError)
		return
	}

	WriteJSON(w, http.StatusCreated, map[string]int{"id": id})
}

func (h *Handler) PatchTask(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 0 {
		WriteError(w, "id is incorrect", http.StatusBadRequest)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<10)

	var body struct {
		Title string `json:"title"`
		Done  bool   `json:"done"`
	}

	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&body); err != nil {
		if _, ok := errors.AsType[*http.MaxBytesError](err); ok {
			WriteError(w, "title size is too large", http.StatusRequestEntityTooLarge)
			return
		}

		WriteError(w, "can't read json", http.StatusBadRequest)
		return
	}

	if err := h.postgresStorage.PatchTask(context.Background(), id, body.Title, body.Done); err != nil {
		WriteError(w, "task does not exist", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) DeleteTaskByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 0 {
		WriteError(w, "id is incorrect", http.StatusBadRequest)
		return
	}

	if err := h.postgresStorage.DeleteTaskByID(context.Background(), id); err != nil {
		WriteError(w, "task does not exist", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
