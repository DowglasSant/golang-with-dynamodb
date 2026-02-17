package handler

import "github.com/dowglassantana/golang-with-dynamodb/internal/model"

// UserResponse e o DTO de saida para respostas HTTP.
// Separa a representacao JSON da entidade de dominio.
type UserResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	CreatedAt string `json:"created_at"`
}

func toUserResponse(u model.User) UserResponse {
	return UserResponse{
		ID:        u.ID,
		Name:      u.Name,
		Email:     u.Email,
		CreatedAt: u.CreatedAt,
	}
}

func toUserResponseList(users []model.User) []UserResponse {
	res := make([]UserResponse, len(users))
	for i, u := range users {
		res[i] = toUserResponse(u)
	}
	return res
}
