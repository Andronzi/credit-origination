package domain

import (
	"errors"
	"time"
)

type ApplicationStatus string

const (
	StatusNew        ApplicationStatus = "NEW"
	StatusScoring    ApplicationStatus = "SCORING"
	StatusEmployment ApplicationStatus = "EMPLOYMENT_CHECK"
	StatusApproved   ApplicationStatus = "APPROVED"
	StatusRejected   ApplicationStatus = "REJECTED"
)

type CreditApplication struct {
	ID           string  `gorm:"primaryKey"`
	UserID       string  `gorm:"index;foreignKey:UserID;references:ID"`
	Amount       float64 `gorm:"type:decimal(15,2)"`
	Term         int
	Status       ApplicationStatus `gorm:"type:varchar(20)"`
	RejectReason string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (a *CreditApplication) Validate() error {
	if a.Amount <= 0 {
		return errors.New("amount must be positive")
	}
	return nil
}
