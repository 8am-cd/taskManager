package repository

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"time"
)

func (s *PostgresStorage) SessionActive(tokenHash string) (int, error) {
	var id int

	err := s.pool.QueryRow(
		context.Background(),
		`SELECT user_id FROM sessions WHERE hash = $1 AND now() < expires_at`,
		tokenHash,
	).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (s *PostgresStorage) addSession(userID int) (string, time.Time, error) {
	var token string
	var tokenHash string
	var expiresAt time.Time
	token, tokenHash, _ = generateSessionToken()

	err := s.pool.QueryRow(
		context.Background(),
		`INSERT INTO sessions (hash, user_id) VALUES ($1, $2) RETURNING expires_at`,
		tokenHash,
		userID,
	).Scan(&expiresAt)

	return token, expiresAt, err
}

func (s *PostgresStorage) deleteSession(tokenHash string) error {
	_, err := s.pool.Exec(
		context.Background(),
		`UPDATE sessions SET expires_at = now() WHERE hash = $1`,
		tokenHash,
	)
	return err
}

func generateSessionToken() (token, tokenHash string, err error) {
	var randomBytes [32]byte
	if _, err := rand.Read(randomBytes[:]); err != nil {
		return "", "", err
	}

	token = base64.RawURLEncoding.EncodeToString(randomBytes[:])
	sum := sha256.Sum256([]byte(token))
	return token, hex.EncodeToString(sum[:]), nil
}
