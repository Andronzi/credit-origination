package domain

import "context"

type CreditRepository interface {
	Save(ctx context.Context, app *CreditApplication) error
	FindByID(ctx context.Context, id string) (*CreditApplication, error)
	FindByUserID(ctx context.Context, userID string) (*CreditApplication, error)
	UpdateStatus(ctx context.Context, id string, status ApplicationStatus) error
}
