package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/dowglassantana/golang-with-dynamodb/internal/handler"
	"github.com/dowglassantana/golang-with-dynamodb/internal/repository"
	"github.com/dowglassantana/golang-with-dynamodb/internal/service"
	"github.com/dowglassantana/golang-with-dynamodb/pkg/dynamo"
)

func main() {
	ctx := context.Background()

	env := os.Getenv("ENV")
	if env == "" {
		env = "local"
	}

	var (
		client *dynamodb.Client
		err    error
	)

	switch env {
	case "aws":
		region := os.Getenv("AWS_REGION")
		if region == "" {
			region = "us-east-1"
		}
		client, err = dynamo.NewClient(ctx, region)
	default:
		client, err = dynamo.NewLocalClient(ctx)
	}

	if err != nil {
		log.Fatalf("erro ao criar client DynamoDB: %v", err)
	}

	repo := repository.NewUserRepository(client)

	if err := repo.CreateTable(ctx); err != nil {
		log.Printf("aviso ao criar tabela (pode ja existir): %v", err)
	}

	svc := service.NewUserService(repo)
	userHandler := handler.NewUserHandler(svc)

	mux := http.NewServeMux()
	userHandler.RegisterRoutes(mux)

	addr := ":8080"
	fmt.Printf("Servidor rodando em http://localhost%s (env=%s)\n", addr, env)
	log.Fatal(http.ListenAndServe(addr, mux))
}
