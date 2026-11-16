package handlers

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"

	"github.com/jackc/pgx/v5"
)

type contextKey int8

const keyUserID contextKey = iota

func (h *Handler) Authorization(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session")
		if err != nil {
			WriteError(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		var userID int
		sum := sha256.Sum256([]byte(cookie.Value))
		userID, err = h.postgresStorage.SessionActive(hex.EncodeToString(sum[:]))
		if errors.Is(err, pgx.ErrNoRows) {
			WriteError(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		if err != nil {
			WriteError(w, "internal server error", http.StatusInternalServerError)
			return
		}

		ctx := context.WithValue(r.Context(), keyUserID, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
