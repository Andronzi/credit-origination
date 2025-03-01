package usecase

import (
	"context"
	"log"

	"github.com/Andronzi/credit-origination/internal/client"
	"github.com/Andronzi/credit-origination/internal/domain"
)

type CreateApplicationUseCase struct {
	repo    domain.CreditRepository
	scoring *client.ScoringClient
}

func NewCreateApplicationUseCase(
	repo domain.CreditRepository,
	scoring *client.ScoringClient,
) *CreateApplicationUseCase {
	return &CreateApplicationUseCase{repo, scoring}
}

func (uc *CreateApplicationUseCase) Execute(ctx context.Context, app *domain.CreditApplication) error {
	log.Printf("Creating application with ID: %s", app.ID)

	if err := uc.repo.Save(ctx, app); err != nil {
		log.Printf("Repository error: %v", err)
		return err
	}

	// TODO: Добавить асинхронное действие верификации
	// go uc.verifyAsync(ctx, app.ID)

	log.Printf("Application created successfully: %s", app.ID)

	return nil
}

// TODO: Реализовать логику верификации заявки
/*
func (uc *CreateApplicationUseCase) verifyAsync(ctx context.Context, appID string) {
    // ... логика с вызовами scoring и antifraud
}
*/
