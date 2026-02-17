package model

// UpdateUserInput e o DTO de entrada para atualizacao de usuario.
type UpdateUserInput struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}
