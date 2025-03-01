package usecase

import (
	"context"
	"errors"

	"github.com/Andronzi/credit-origination/internal/domain"
)

type ListApplicationResult struct {
	Applications []*domain.CreditApplication
	CurrentPage  int
	PageSize     int
	TotalCount   int
	TotalPages   int
}

type ListApplicationUseCase struct {
	repo domain.CreditRepository
}

func NewListApplicationUseCase(repo domain.CreditRepository) *ListApplicationUseCase {
	return &ListApplicationUseCase{
		repo,
	}
}

func (uc *ListApplicationUseCase) Execute(ctx context.Context, statuses []domain.ApplicationStatus, page int, pageSize int) (*ListApplicationResult, error) {
	if page <= 0 || pageSize <= 0 {
		return nil, errors.New("invalid pagination parameters")
	}

	offset := (page - 1) * pageSize

	applications, totalCount, err := uc.repo.List(ctx, statuses, offset, pageSize)
	if err != nil {
		return nil, err
	}

	totalPages := (totalCount + pageSize - 1) / pageSize

	return &ListApplicationResult{
		Applications: applications,
		CurrentPage:  page,
		PageSize:     pageSize,
		TotalCount:   totalCount,
		TotalPages:   totalPages,
	}, nil
}
