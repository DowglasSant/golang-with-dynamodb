package model

// CreateUserInput e o DTO de entrada para criacao de usuario.
type CreateUserInput struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}
