package repository

import "github.com/dowglassantana/golang-with-dynamodb/internal/model"

// userDynamo e a representacao do usuario no DynamoDB.
// As tags `dynamodbav` mapeiam os campos para os atributos da tabela.
// Esse model existe apenas na camada de repository — o restante da aplicacao
// trabalha com model.User, que nao conhece DynamoDB.
type userDynamo struct {
	ID        string `dynamodbav:"id"`
	Name      string `dynamodbav:"name"`
	Email     string `dynamodbav:"email"`
	CreatedAt string `dynamodbav:"created_at"`
}

// toDynamo converte model.User (dominio) para userDynamo (DynamoDB).
func toDynamo(u model.User) userDynamo {
	return userDynamo{
		ID:        u.ID,
		Name:      u.Name,
		Email:     u.Email,
		CreatedAt: u.CreatedAt,
	}
}

// toUser converte userDynamo (DynamoDB) para model.User (dominio).
func (m userDynamo) toUser() model.User {
	return model.User{
		ID:        m.ID,
		Name:      m.Name,
		Email:     m.Email,
		CreatedAt: m.CreatedAt,
	}
}
