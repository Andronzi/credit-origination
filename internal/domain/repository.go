package domain

import (
	"context"
)

type CreditRepository interface {
	FindByID(ctx context.Context, id string) (*CreditApplication, error)
	FindByUserID(ctx context.Context, userID string) (*CreditApplication, error)
	List(ctx context.Context, statuses []ApplicationStatus, offset int, limit int, userID string) ([]*CreditApplication, int, error)
	Save(ctx context.Context, app *CreditApplication) error
	Update(ctx context.Context, app *CreditApplication) error
	UpdateStatus(ctx context.Context, id string, status ApplicationStatus) error
	Delete(ctx context.Context, id string) error
}
