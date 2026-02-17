package model

import (
	"time"

	"github.com/google/uuid"
)

// User e a entidade de dominio — representa um usuario na aplicacao.
type User struct {
	ID        string
	Name      string
	Email     string
	CreatedAt string
}

func NewUser(name, email string) User {
	return User{
		ID:        uuid.New().String(),
		Name:      name,
		Email:     email,
		CreatedAt: time.Now().Format(time.RFC3339),
	}
}
