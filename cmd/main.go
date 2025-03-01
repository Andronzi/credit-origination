package main

import (
	"log"
	"net"

	"github.com/Andronzi/credit-origination/internal/client"
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

	creditRepo := repository.NewCreditRepo(db)

	scoringClient := client.NewScoringClient("http://scoring-service:8080")

	createApplicationUC := usecase.NewCreateApplicationUseCase(
		creditRepo,
		scoringClient,
	)
	listApplicationUC := usecase.NewListApplicationUseCase(creditRepo)

	grpcServer := grpc.NewServer()
	createApplicationServer := grpcserver.NewCreateApplicationServer(createApplicationUC, listApplicationUC)

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
