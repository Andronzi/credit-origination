package repository

import (
	"context"
	"log"

	"github.com/Andronzi/credit-origination/internal/domain"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type CreditRepo struct {
	db *gorm.DB
}

// TODO: Подумать можно ли лучше расположить проверку имплементации
var _ domain.CreditRepository = (*CreditRepo)(nil)

func NewCreditRepo(db *gorm.DB) *CreditRepo {
	db.Logger = db.Logger.LogMode(logger.Info)
	return &CreditRepo{db: db}
}

func (r *CreditRepo) Save(ctx context.Context, app *domain.CreditApplication) error {
	log.Printf("Attempting to save application: %+v", app)
	return r.db.WithContext(ctx).Create(app).Error
}

func (r *CreditRepo) FindByID(ctx context.Context, id string) (*domain.CreditApplication, error) {
	var app domain.CreditApplication
	err := r.db.WithContext(ctx).First(&app, "id = ?", id).Error
	if err != nil {
		log.Printf("Error saving application: %v", err)
	}
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

func (r *CreditRepo) Update(ctx context.Context, app *domain.CreditApplication) error {
	return r.db.WithContext(ctx).Model(&domain.CreditApplication{}).
		Where("id = ?", app.ID).
		Updates(app).Error
}

func (r *CreditRepo) Delete(ctx context.Context, appID string) error {
	return r.db.WithContext(ctx).Where("id = ?", appID).Delete(&domain.CreditApplication{}).Error
}

func (r *CreditRepo) List(ctx context.Context, statuses []domain.ApplicationStatus, offset int, limit int) ([]*domain.CreditApplication, int, error) {
	var applications []*domain.CreditApplication

	query := r.db.WithContext(ctx).Model(&domain.CreditApplication{})
	if len(statuses) > 0 {
		query = query.Where("status IN ?", statuses)
	}

	var totalCount int64
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Offset(offset).Limit(limit).Find(&applications).Error; err != nil {
		return nil, 0, err
	}

	return applications, int(totalCount), nil
}
