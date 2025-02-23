package usecase

import (
	"context"

	"github.com/Andronzi/credit-origination/internal/domain"
)

type GetApplicationUseCase struct {
	repo domain.CreditRepository
}

func NewGetApplicationUseCase(
	repo domain.CreditRepository,
) *GetApplicationUseCase {
	return &GetApplicationUseCase{repo}
}

func (uc *GetApplicationUseCase) Execute(ctx context.Context, appID string) (*domain.CreditApplication, error) {
	return uc.repo.FindByID(ctx, appID)
}
