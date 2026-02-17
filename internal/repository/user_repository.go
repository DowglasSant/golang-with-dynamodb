package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/dowglassantana/golang-with-dynamodb/internal/entity"
)

const tableName = "Users"

type UserRepository struct {
	client *dynamodb.Client
}

func NewUserRepository(client *dynamodb.Client) *UserRepository {
	return &UserRepository{client: client}
}

// CreateTable cria a tabela "Users" no DynamoDB caso ela ainda nao exista.
//
// No DynamoDB, toda tabela precisa de pelo menos uma chave primaria (partition key).
// Aqui usamos o campo "id" como partition key (HASH), o que significa que cada item
// na tabela sera identificado unicamente pelo seu "id".
//
// KeySchema define a estrutura da chave:
//   - HASH = partition key (obrigatoria) — distribui os dados entre as particoes internas.
//
// AttributeDefinitions descreve o tipo do atributo usado na chave:
//   - "S" = String, "N" = Number, "B" = Binary.
//
// BillingMode PAY_PER_REQUEST = modo sob demanda (sem necessidade de provisionar capacidade).
// Ideal para desenvolvimento local e cargas imprevisiveis.
func (r *UserRepository) CreateTable(ctx context.Context) error {
	_, err := r.client.CreateTable(ctx, &dynamodb.CreateTableInput{
		TableName: aws.String(tableName),
		KeySchema: []types.KeySchemaElement{
			{
				AttributeName: aws.String("id"),
				KeyType:       types.KeyTypeHash,
			},
		},
		AttributeDefinitions: []types.AttributeDefinition{
			{
				AttributeName: aws.String("id"),
				AttributeType: types.ScalarAttributeTypeS,
			},
		},
		BillingMode: types.BillingModePayPerRequest,
	})
	if err != nil {
		// ResourceInUseException significa que a tabela ja existe — podemos ignorar.
		var resourceInUse *types.ResourceInUseException
		if errors.As(err, &resourceInUse) {
			return nil
		}
		return fmt.Errorf("erro ao criar tabela: %w", err)
	}
	return nil
}

// Create insere um novo usuario na tabela usando PutItem.
//
// PutItem e a operacao basica de escrita do DynamoDB. Ela insere um item novo
// ou substitui completamente um item existente que tenha a mesma chave primaria.
//
// attributevalue.MarshalMap converte a struct Go para o formato map[string]AttributeValue
// que o DynamoDB espera. Ele usa as tags `dynamodbav` da struct para mapear os campos.
//
// Exemplo: entity.User{ID: "123", Name: "Joao"} vira:
//
//	map[string]AttributeValue{
//	    "id":   &types.AttributeValueMemberS{Value: "123"},
//	    "name": &types.AttributeValueMemberS{Value: "Joao"},
//	}
func (r *UserRepository) Create(ctx context.Context, user entity.User) error {
	item, err := attributevalue.MarshalMap(user)
	if err != nil {
		return fmt.Errorf("erro ao serializar usuario: %w", err)
	}

	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(tableName),
		Item:      item,
	})
	if err != nil {
		return fmt.Errorf("erro ao inserir usuario: %w", err)
	}

	return nil
}

// GetByID busca um usuario pelo ID usando GetItem.
//
// GetItem e a forma mais eficiente de ler um unico item no DynamoDB,
// pois acessa diretamente a particao correta usando a chave primaria (partition key).
// A complexidade e O(1) — nao importa quantos itens existam na tabela.
//
// O parametro Key recebe a chave primaria do item que queremos buscar.
// Como nossa partition key e "id" do tipo String, passamos um AttributeValueMemberS.
//
// Se o item nao for encontrado, GetItem retorna sem erro, mas output.Item vem nil.
// Por isso verificamos se o resultado esta vazio antes de tentar desserializar.
//
// attributevalue.UnmarshalMap faz o caminho inverso do MarshalMap:
// converte o map[string]AttributeValue de volta para a struct Go.
func (r *UserRepository) GetByID(ctx context.Context, id string) (*entity.User, error) {
	output, err := r.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: id},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar usuario: %w", err)
	}

	if output.Item == nil {
		return nil, nil
	}

	var user entity.User
	err = attributevalue.UnmarshalMap(output.Item, &user)
	if err != nil {
		return nil, fmt.Errorf("erro ao desserializar usuario: %w", err)
	}

	return &user, nil
}

// GetAll retorna todos os usuarios da tabela usando Scan.
//
// Scan percorre TODOS os itens da tabela e retorna cada um deles.
// Diferente de GetItem (que busca por chave), o Scan faz uma varredura completa.
//
// ATENCAO: Scan e uma operacao custosa! Ele le cada item da tabela, o que consome
// muitas unidades de leitura (RCU). Em tabelas grandes, isso pode ser lento e caro.
// Para producao, prefira usar Query com indices (GSI/LSI) quando possivel.
//
// Scan retorna no maximo 1MB de dados por chamada. Para tabelas maiores,
// seria necessario usar paginacao com LastEvaluatedKey, mas para este exemplo
// basico uma unica chamada e suficiente.
//
// attributevalue.UnmarshalListOfMaps converte a lista de items retornada pelo
// DynamoDB para um slice de structs Go.
func (r *UserRepository) GetAll(ctx context.Context) ([]entity.User, error) {
	output, err := r.client.Scan(ctx, &dynamodb.ScanInput{
		TableName: aws.String(tableName),
	})
	if err != nil {
		return nil, fmt.Errorf("erro ao listar usuarios: %w", err)
	}

	var users []entity.User
	err = attributevalue.UnmarshalListOfMaps(output.Items, &users)
	if err != nil {
		return nil, fmt.Errorf("erro ao desserializar usuarios: %w", err)
	}

	return users, nil
}

// Update atualiza os campos name e email de um usuario usando UpdateItem.
//
// UpdateItem modifica atributos especificos de um item existente SEM substituir
// o item inteiro (diferente de PutItem que sobrescreve tudo).
// Isso e mais eficiente quando queremos alterar apenas alguns campos.
//
// Usamos o pacote expression para construir a UpdateExpression de forma segura.
// Ele gera automaticamente:
//   - UpdateExpression: "SET #name = :name, #email = :email"
//   - ExpressionAttributeNames: {"#name": "name", "#email": "email"}
//   - ExpressionAttributeValues: {":name": "Joao", ":email": "joao@email.com"}
//
// Por que usar ExpressionAttributeNames (#name)?
// Porque "name" e uma palavra reservada do DynamoDB. Se usarmos "name" diretamente
// na expressao, o DynamoDB retorna erro. O alias #name evita esse conflito.
//
// Por que usar ExpressionAttributeValues (:name)?
// Para evitar injecao de expressoes e separar a logica dos valores, similar
// a prepared statements em SQL.
//
// ConditionExpression "attribute_exists(id)" garante que so atualizamos um item
// que ja existe. Se o id nao for encontrado, o DynamoDB retorna ConditionalCheckFailedException.
func (r *UserRepository) Update(ctx context.Context, id string, input entity.UpdateUserInput) error {
	update := expression.
		Set(expression.Name("name"), expression.Value(input.Name)).
		Set(expression.Name("email"), expression.Value(input.Email))

	condition := expression.AttributeExists(expression.Name("id"))

	expr, err := expression.NewBuilder().
		WithUpdate(update).
		WithCondition(condition).
		Build()
	if err != nil {
		return fmt.Errorf("erro ao construir expressao: %w", err)
	}

	_, err = r.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(tableName),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: id},
		},
		UpdateExpression:          expr.Update(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		ConditionExpression:       expr.Condition(),
	})
	if err != nil {
		return fmt.Errorf("erro ao atualizar usuario: %w", err)
	}

	return nil
}

// Delete remove um usuario da tabela pelo ID usando DeleteItem.
//
// DeleteItem remove um unico item com base na chave primaria informada.
// Assim como GetItem, a operacao acessa diretamente a particao correta.
//
// Por padrao, DeleteItem NAO retorna erro se o item nao existir — ele simplesmente
// nao faz nada (operacao idempotente). Isso e util para garantir que
// chamadas repetidas de delete nao causem falhas.
//
// Se quisermos garantir que o item existia antes de deletar, podemos adicionar
// ConditionExpression: aws.String("attribute_exists(id)"), que faria o DynamoDB
// retornar erro caso o item nao fosse encontrado.
func (r *UserRepository) Delete(ctx context.Context, id string) error {
	_, err := r.client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(tableName),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: id},
		},
	})
	if err != nil {
		return fmt.Errorf("erro ao deletar usuario: %w", err)
	}

	return nil
}
