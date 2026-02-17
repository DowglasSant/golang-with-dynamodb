package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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

	tableName := os.Getenv("DYNAMO_TABLE")
	if tableName == "" {
		tableName = "Users"
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

	repo := repository.NewUserRepository(client, tableName)

	if err := repo.CreateTable(ctx); err != nil {
		log.Printf("aviso ao criar tabela (pode ja existir): %v", err)
	}

	svc := service.NewUserService(repo)
	userHandler := handler.NewUserHandler(svc)

	mux := http.NewServeMux()
	userHandler.RegisterRoutes(mux)

	addr := ":8080"
	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	// Graceful shutdown: ao receber SIGINT ou SIGTERM, o servidor para de aceitar
	// novas conexoes e aguarda ate 10 segundos para as requests em andamento finalizarem.
	// No ECS Fargate, o container recebe SIGTERM antes de ser encerrado.
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		sig := <-sigChan
		log.Printf("sinal recebido: %v. Encerrando servidor...", sig)

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("erro ao encerrar servidor: %v", err)
		}
	}()

	fmt.Printf("Servidor rodando em http://localhost%s (env=%s)\n", addr, env)
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("erro no servidor: %v", err)
	}

	log.Println("servidor encerrado com sucesso")
}
