package domain

import (
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// ApplicationStatus представляет статус заявки
// @Schema(description="Статус кредитной заявки", enum=["NEW","SCORING","EMPLOYMENT_CHECK","APPROVED","REJECTED"])
type ApplicationStatus int32

const (
	NEW              ApplicationStatus = 0
	SCORING          ApplicationStatus = 1
	EMPLOYMENT_CHECK ApplicationStatus = 2
	APPROVED         ApplicationStatus = 3
	REJECTED         ApplicationStatus = 4
)

// CreditApplication представляет кредитную заявку
// @Schema(description="Доменная модель кредитной заявки")
type CreditApplication struct {
	ID           uuid.UUID         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	UserID       uuid.UUID         `gorm:"type:uuid; json:"user_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Amount       decimal.Decimal   `gorm:"type:decimal(15,2)" json:"amount" example:"150000.50"`
	Term         uint32            `json:"term" example:"12"`
	Interest     decimal.Decimal   `json:"interest" example:"15"`
	Status       ApplicationStatus `gorm:"type:int32"; "default: 0"`
	RejectReason sql.NullString    `json:"reject_reason" example:"Low credit score"`
	CreatedAt    time.Time         `json:"created_at" example:"2023-10-01T12:34:56Z"`
	UpdatedAt    time.Time         `json:"updated_at" example:"2023-10-01T12:34:56Z"`
	// TODO: Добавить FK UserID, когда будет интеграция с пользователями :)
	// UserID       uuid.UUID         `gorm:"type:uuid;index;foreignKey:UserID;references:ID" json:"user_id" example:"550e8400-e29b-41d4-a716-446655440000"`
}

func NewCreditApplication(amount decimal.Decimal, interest decimal.Decimal, term uint32) (*CreditApplication, error) {
	if amount.CoefficientInt64() <= 0 {
		return nil, errors.New("amount must be positive")
	}

	if interest.CoefficientInt64() <= 0 {
		return nil, errors.New("interest must be positive")
	}

	if term <= 0 {
		return nil, errors.New("term must be positive")
	}

	now := time.Now().UTC()
	return &CreditApplication{
		ID:        uuid.New(),
		Amount:    amount,
		Term:      term,
		Interest:  interest,
		Status:    ApplicationStatus(SCORING),
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}
