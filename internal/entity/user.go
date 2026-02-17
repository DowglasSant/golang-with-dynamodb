package entity

import "time"

type User struct {
	ID        string `json:"id" dynamodbav:"id"`
	Name      string `json:"name" dynamodbav:"name"`
	Email     string `json:"email" dynamodbav:"email"`
	CreatedAt string `json:"created_at" dynamodbav:"created_at"`
}

type CreateUserInput struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type UpdateUserInput struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

func NewUser(name, email string) User {
	return User{
		ID:        "",
		Name:      name,
		Email:     email,
		CreatedAt: time.Now().Format(time.RFC3339),
	}
}
