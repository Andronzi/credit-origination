package main

import (
	"github.com/Andronzi/credit-origination/internal/client"
	"github.com/Andronzi/credit-origination/internal/handlers"
	"github.com/Andronzi/credit-origination/internal/repository"
	"github.com/Andronzi/credit-origination/internal/usecase"
	"github.com/gin-gonic/gin"
)

func main() {
	// TODO: Добавить logger
	// logger := logger.New()

	// TODO: Добавить database
	// db := database.ConnectPostgres()
	// db.AutoMigrate(&domain.CreditApplication{})

	creditRepo := repository.NewCreditRepo(db)

	scoringClient := client.NewScoringClient("http://scoring-service:8080")

	createApplicationUC := usecase.NewCreateApplicationUseCase(
		creditRepo,
		scoringClient,
	)
	getApplicationUC := usecase.NewGetApplicationUseCase(creditRepo)

	applicationHandler := handlers.NewCreateCreditHandler(
		createApplicationUC,
		getApplicationUC,
	)

	router := gin.Default()
	router.POST("/application", applicationHandler.CreateApplication)
	router.Run(":8080")
}
