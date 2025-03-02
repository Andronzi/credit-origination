package main

import (
	"fmt"
	"log"
	"net"
	"os"

	"github.com/Andronzi/credit-origination/internal/client"
	"github.com/Andronzi/credit-origination/internal/messaging"
	"github.com/Andronzi/credit-origination/internal/repository"
	grpcserver "github.com/Andronzi/credit-origination/internal/transport/grpc"
	"github.com/Andronzi/credit-origination/internal/usecase"
	"github.com/Andronzi/credit-origination/pkg/database"
	"github.com/Andronzi/credit-origination/pkg/grpc/credit"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	// TODO: Добавить полноценный logger
	log.Printf("Старт прилы")

	db, err := database.ConnectPostgres()
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}
	log.Printf("Успешный Connect к базе")

	kafkaProducer, err := initKafkaProducer()
	if err != nil {
		log.Fatalf("Failed to init Kafka producer: %v", err)
	}

	creditRepo := repository.NewCreditRepo(db)

	scoringClient := client.NewScoringClient("http://scoring-service:8080")

	createApplicationUC := usecase.NewCreateApplicationUseCase(
		creditRepo,
		scoringClient,
	)
	listApplicationUC := usecase.NewListApplicationUseCase(creditRepo)
	getApplicationUC := usecase.NewGetApplicationUseCase(creditRepo)
	updateApplicationUC := usecase.NewUpdateApplicationUseCase(creditRepo)
	deleteApplicationUC := usecase.NewDeleteApplicationUseCase(creditRepo)

	grpcServer := grpc.NewServer()
	createApplicationServer := grpcserver.NewCreateApplicationServer(
		getApplicationUC,
		createApplicationUC,
		listApplicationUC,
		updateApplicationUC,
		deleteApplicationUC,
		kafkaProducer,
	)

	credit.RegisterApplicationServiceServer(grpcServer, createApplicationServer)

	reflection.Register(grpcServer)

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	log.Println("gRPC server is running on port 50051")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

// TODO: Унифицировать создание
func initKafkaProducer() (*messaging.KafkaProducer, error) {
	schema, err := os.ReadFile("/schemas/avro/credit/v1/StatusEvent.avsc")
	if err != nil {
		return nil, fmt.Errorf("failed to read schema file: %w", err)
	}

	return messaging.NewKafkaProducer(
		[]string{"kafka:9092"},
		"application",
		string(schema),
	)
}
