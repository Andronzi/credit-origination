package repository

import (
	"context"

	"github.com/Andronzi/credit-origination/internal/domain"
	"gorm.io/gorm"
)

type CreditRepo struct {
	db *gorm.DB
}

// TODO: Подумать можно ли лучше расположить проверку имплементации
var _ domain.CreditRepository = (*CreditRepo)(nil)

func NewCreditRepo(db *gorm.DB) *CreditRepo {
	return &CreditRepo{db: db}
}

func (r *CreditRepo) Save(ctx context.Context, app *domain.CreditApplication) error {
	return r.db.WithContext(ctx).Create(app).Error
}

func (r *CreditRepo) FindByID(ctx context.Context, id string) (*domain.CreditApplication, error) {
	var app domain.CreditApplication
	err := r.db.WithContext(ctx).First(&app, "id = ?", id).Error
	return &app, err
}

func (r *CreditRepo) FindByUserID(ctx context.Context, userID string) (*domain.CreditApplication, error) {
	var app domain.CreditApplication
	err := r.db.WithContext(ctx).First(&app, "userID = ?", userID).Error
	return &app, err
}

func (r *CreditRepo) UpdateStatus(ctx context.Context, id string, status domain.ApplicationStatus) error {
	return r.db.WithContext(ctx).
		Model(&domain.CreditApplication{}).
		Where("id = ?", id).
		Update("status", status).Error
}
