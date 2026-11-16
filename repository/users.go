package repository

import (
	"awesomeProject/users"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
)

func (s *PostgresStorage) AddUser(user *users.User) error {
	userExist, err := s.UserExists(user.Email)
	if err != nil {
		log.Printf("failed query to find by email: %v", err)
		return err
	}
	if userExist == true {
		log.Printf("user already exists")
		return &ErrUserExists{Email: user.Email}
	}

	password_hash, _ := HashPassword(user.Password)

	if _, err := s.pool.Exec(
		context.Background(),
		`INSERT INTO users (email, password_hash) VALUES ($1, $2);`,
		user.Email,
		password_hash,
	); err != nil {
		log.Printf("failed query to add user")
		return err
	}

	return nil
}

func (s *PostgresStorage) UserExists(email string) (bool, error) {
	var id int
	var emailFind string

	err := s.pool.QueryRow(
		context.Background(),
		`SELECT id, email FROM users WHERE email = $1`,
		email,
	).Scan(&id, &emailFind)

	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return true, nil
}

func (s *PostgresStorage) LoginUser(user *users.User) (string, time.Time, error) {
	var id int
	var passwordFind string
	err := s.pool.QueryRow(
		context.Background(),
		`SELECT id, password_hash 
			 FROM users 
			 WHERE email = $1`,
		user.Email,
	).Scan(&id, &passwordFind)

	if errors.Is(err, pgx.ErrNoRows) {
		return "", time.Time{}, &ErrUserNotExists{Email: user.Email}
	}
	if err != nil {
		return "", time.Time{}, err
	}
	if !CheckPasswordHash(user.Password, passwordFind) {
		return "", time.Time{}, &ErrWrongPassword{}
	}

	var token string
	var expiresAt time.Time
	token, expiresAt, err = s.addSession(id)
	if err != nil {
		return "", time.Time{}, err
	}

	return token, expiresAt, nil
}

func (s *PostgresStorage) LogoutUser(token string) error {
	sum := sha256.Sum256([]byte(token))
	err := s.deleteSession(hex.EncodeToString(sum[:]))
	return err
}
