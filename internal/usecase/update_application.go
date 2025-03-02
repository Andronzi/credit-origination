package usecase

import (
	"context"

	"github.com/Andronzi/credit-origination/internal/domain"
)

type UpdateApplicationUseCase struct {
	repo domain.CreditRepository
}

func NewUpdateApplicationUseCase(
	repo domain.CreditRepository,
) *UpdateApplicationUseCase {
	return &UpdateApplicationUseCase{repo}
}

func (uc *UpdateApplicationUseCase) Execute(ctx context.Context, app *domain.CreditApplication) error {
	return uc.repo.Update(ctx, app)
}
