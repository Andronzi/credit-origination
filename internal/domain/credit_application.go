package domain

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type ApplicationStatus string

const (
	DRAFT            ApplicationStatus = "DRAFT"
	APPLICATION      ApplicationStatus = "APPLICATION"
	SCORING          ApplicationStatus = "SCORING"
	EMPLOYMENT_CHECK ApplicationStatus = "EMPLOYMENT_CHECK"
	APPROVED         ApplicationStatus = "APPROVED"
	REJECTED         ApplicationStatus = "REJECTED"
)

type CreditApplication struct {
	ID                 uuid.UUID         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	DisbursementAmount decimal.Decimal   `gorm:"type:decimal(15,2)" json:"disbursement_amount" example:"150000.50"`
	OriginationAmount  decimal.Decimal   `gorm:"type:decimal(15,2)" json:"origination_amount" example:"150000.50"`
	ToBankAccountID    uuid.UUID         `gorm:"type:uuid;" json:"account_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	UserID             uuid.UUID         `gorm:"type:uuid;index;foreignKey:UserID;references:ID" json:"user_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Term               uint32            `json:"term" example:"12"`
	Interest           decimal.Decimal   `json:"interest" example:"15"`
	ProductCode        string            `gorm:"type:string; json:"product_code" example:"code-1"`
	ProductVersion     string            `gorm:"type:srting; json:"product_version" example:"version1"`
	Status             ApplicationStatus `gorm:"type:string"; "default: "DRAFT"`
	RejectReason       sql.NullString    `json:"reject_reason" example:"Low credit score"`
	CreatedAt          time.Time         `json:"created_at" example:"2023-10-01T12:34:56Z"`
	UpdatedAt          time.Time         `json:"updated_at" example:"2023-10-01T12:34:56Z"`
}

var (
	ErrInvalidAmount             = errors.New("invalid amount")
	ErrEmptyBankAccount          = errors.New("bank account ID is required")
	ErrInvalidProduct            = errors.New("invalid product")
	ErrInvalidTerm               = errors.New("term must be greater than zero")
	ErrInvalidInterest           = errors.New("interest rate must be positive")
	ErrInvalidUser               = errors.New("user ID is required")
	ErrInvalidProductCode        = errors.New("product code is required")
	ErrInvalidProductVersion     = errors.New("product version cannot be empty")
	ErrDisbursementExceedsAmount = errors.New("disbursement cannot exceed application amount")
)

func NewCreditApplication(
	disbursementAmount decimal.Decimal,
	originationAmount decimal.Decimal,
	toBankAccountID uuid.UUID,
	term uint32,
	interest decimal.Decimal,
	productCode string,
	productVersion string,
	userID uuid.UUID,
) (*CreditApplication, error) {
	if disbursementAmount.IsZero() || disbursementAmount.IsNegative() {
		return nil, fmt.Errorf("%w: disbursement amount must be positive", ErrInvalidAmount)
	}

	if originationAmount.IsZero() || originationAmount.IsNegative() {
		return nil, fmt.Errorf("%w: origination amount must be positive", ErrInvalidAmount)
	}

	if originationAmount.GreaterThan(disbursementAmount) {
		return nil, ErrDisbursementExceedsAmount
	}

	if term == 0 {
		return nil, ErrInvalidTerm
	}

	if interest.IsZero() || interest.IsNegative() {
		return nil, ErrInvalidInterest
	}

	if err := validateUUID(toBankAccountID, "bank account ID"); err != nil {
		return nil, err
	}

	if err := validateUUID(userID, "user ID"); err != nil {
		return nil, err
	}

	productVersion = strings.TrimSpace(productVersion)
	if productVersion == "" {
		return nil, ErrInvalidProductVersion
	}

	productVersion = strings.ToLower(productVersion)

	now := time.Now().UTC()
	return &CreditApplication{
		ID:                 uuid.New(),
		DisbursementAmount: disbursementAmount,
		OriginationAmount:  originationAmount,
		ToBankAccountID:    toBankAccountID,
		UserID:             userID,
		Term:               term,
		Interest:           interest,
		ProductCode:        productCode,
		ProductVersion:     productVersion,
		Status:             DRAFT,
		CreatedAt:          now,
		UpdatedAt:          now,
	}, nil
}

func validateUUID(id uuid.UUID, fieldName string) error {
	if id == uuid.Nil {
		return fmt.Errorf("%w: %s is required", ErrEmptyBankAccount, fieldName)
	}
	return nil
}
