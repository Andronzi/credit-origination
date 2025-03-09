package main

import (
	"context"
	"log"
	"net"
	"os"
	"time"

	"github.com/Andronzi/credit-origination/internal/client"
	"github.com/Andronzi/credit-origination/internal/messaging"
	"github.com/Andronzi/credit-origination/internal/messaging/handlers"
	"github.com/Andronzi/credit-origination/internal/repository"
	grpcserver "github.com/Andronzi/credit-origination/internal/transport/grpc"
	"github.com/Andronzi/credit-origination/internal/usecase"
	"github.com/Andronzi/credit-origination/pkg/database"
	"github.com/Andronzi/credit-origination/pkg/grpc/credit"
	"github.com/Andronzi/credit-origination/pkg/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	logger.InitLogger()
	defer logger.Logger.Sync()

	testFile, err := os.OpenFile("/var/log/myapp.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("Can not work with log file: %v", err)
		// TODO: Обдумать тщатильнее данный момент
		os.Exit(1)
	}
	testFile.Close()

	logger.Logger.Info("Starting application", zap.String("origination-service", "main.go"))

	db, err := database.ConnectPostgres()
	if err != nil {
		logger.Logger.Fatal("Database connection failed", zap.Error(err))
	}
	logger.Logger.Info("Database connection success", zap.String("origination-service", "main.go"))

	conn, err := net.DialTimeout("tcp", "host.docker.internal:8081", 5*time.Second)
	if err != nil {
		logger.Logger.Fatal("schema registry unavailable: %v", zap.Error(err))
	}
	conn.Close()

	kafkaProducer, err := initKafkaProducer()
	if err != nil {
		logger.Logger.Fatal("Failed to init Kafka producer: %v", zap.Error(err))
	}
	logger.Logger.Info("Kafka producer connection success", zap.String("origination-service", "main.go"))

	creditRepo := repository.NewCreditRepo(db)

	scoringClient := client.NewScoringClient("http://scoring-service:8080")

	createApplicationUC := usecase.NewCreateApplicationUseCase(
		creditRepo,
		scoringClient,
	)
	listApplicationUC := usecase.NewListApplicationUseCase(creditRepo)
	getApplicationUC := usecase.NewGetApplicationUseCase(creditRepo)
	updateApplicationUC := usecase.NewUpdateApplicationUseCase(creditRepo)
	updateStatusUC := usecase.NewUpdateStatusUseCase(creditRepo, kafkaProducer)
	deleteApplicationUC := usecase.NewDeleteApplicationUseCase(creditRepo)

	consumer, err := initKafkaConsumer(creditRepo, updateStatusUC)
	if err != nil {
		logger.Logger.Fatal("Failed to init Kafka consumer: %v", zap.Error(err))
	}
	logger.Logger.Info("Kafka consumer connection success", zap.String("origination-service", "main.go"))

	// TODO: Подумать как лучше горутинки организовать
	go func() {
		for {
			err := consumer.Consume(context.Background())
			if err != nil {
				logger.Logger.Error("Failed to consume message", zap.Error(err))
			}
		}
	}()

	grpcServer := grpc.NewServer()
	createApplicationServer := grpcserver.NewCreateApplicationServer(
		getApplicationUC,
		createApplicationUC,
		listApplicationUC,
		updateApplicationUC,
		updateStatusUC,
		deleteApplicationUC,
		kafkaProducer,
	)

	credit.RegisterApplicationServiceServer(grpcServer, createApplicationServer)

	reflection.Register(grpcServer)

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		logger.Logger.Fatal("failed to listen tcp:", zap.Error(err))
	}

	if err := grpcServer.Serve(lis); err != nil {
		logger.Logger.Fatal("failed to serve gRPC server:", zap.Error(err))
	}

	logger.Logger.Info("gRPC server is running on port 50051")
}

// TODO: Унифицировать создание
func initKafkaProducer() (*messaging.KafkaProducer, error) {
	schema, err := os.ReadFile("/schemas/avro/application/v1/ApplicationEvent.avsc")
	if err != nil {
		return nil, err
	}

	return messaging.NewKafkaProducer(
		[]string{"host.docker.internal:9092"},
		"application",
		string(schema),
	)
}

func initKafkaConsumer(
	creditRepo *repository.CreditRepo,
	updateStatusUC *usecase.UpdateStatusUseCase,
) (*messaging.KafkaAvroConsumer, error) {
	schema, err := os.ReadFile("/schemas/avro/application/v1/ApplicationEvent.avsc")
	if err != nil {
		return nil, err
	}

	agreementHandler, err := handlers.NewAgreementCreatedHandler(updateStatusUC, string(schema))
	if err != nil {
		return nil, err
	}
	scoringHandler, err := handlers.NewScoringHandler(updateStatusUC, string(schema))
	if err != nil {
		return nil, err
	}

	handlers := []messaging.MessageHandler{
		agreementHandler,
		scoringHandler,
	}

	consumer, err := messaging.NewKafkaAvroConsumer(
		[]string{"host.docker.internal:9092"},
		"credit-group",
		"application",
		string(schema),
		handlers,
	)
	if err != nil {
		return nil, err
	}

	return consumer, nil
}
