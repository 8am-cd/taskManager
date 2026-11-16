package handlers

import (
	"awesomeProject/repository"
	"awesomeProject/users"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"
)

func (h *Handler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	user, err := readUser(w, r.Body)
	if err != nil {
		return
	}

	if err := h.postgresStorage.AddUser(user); err != nil {
		if _, ok := errors.AsType[*repository.ErrUserExists](err); ok {
			WriteError(w, "user with this email already exists", http.StatusBadRequest)
		} else {
			WriteError(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *Handler) LoginUser(w http.ResponseWriter, r *http.Request) {
	user, err := readUser(w, r.Body)
	if err != nil {
		return
	}

	var token string
	var expiresAt time.Time
	token, expiresAt, err = h.postgresStorage.LoginUser(user)

	var unknownUser *repository.ErrUserNotExists
	var wrongPassword *repository.ErrWrongPassword

	if errors.As(err, &unknownUser) || errors.As(err, &wrongPassword) {
		WriteError(w, "wrong email or password", http.StatusUnauthorized)
		return
	}
	if err != nil {
		WriteError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    token,
		Path:     "/",
		Expires:  expiresAt,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) LogoutUser(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session")
	if err != nil {
		WriteError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	if err := h.postgresStorage.LogoutUser(cookie.Value); err != nil {
		WriteError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func readUser(w http.ResponseWriter, body io.ReadCloser) (*users.User, error) {
	var user users.User

	dec := json.NewDecoder(body)
	if err := dec.Decode(&user); err != nil {
		WriteError(w, "bad request", http.StatusBadRequest)
		return &users.User{}, err
	}

	user.Email = strings.TrimSpace(user.Email)
	if len(user.Email) == 0 {
		WriteError(w, "email is empty", http.StatusBadRequest)
		return &users.User{}, errors.New("email is empty")
	}

	user.Password = strings.TrimSpace(user.Password)
	if len(user.Password) == 0 {
		WriteError(w, "password is empty", http.StatusBadRequest)
		return &users.User{}, errors.New("password is empty")
	}

	return &user, nil
}
