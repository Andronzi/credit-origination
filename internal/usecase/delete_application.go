package usecase

import (
	"context"

	"github.com/Andronzi/credit-origination/internal/domain"
)

type DeleteApplicationUseCase struct {
	repo domain.CreditRepository
}

func NewDeleteApplicationUseCase(
	repo domain.CreditRepository,
) *DeleteApplicationUseCase {
	return &DeleteApplicationUseCase{repo}
}

func (uc *DeleteApplicationUseCase) Execute(ctx context.Context, appID string) error {
	return uc.repo.Delete(ctx, appID)
}
