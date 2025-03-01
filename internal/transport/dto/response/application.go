package response

import "time"

type CreditApplicationResponse struct {
	ID        string    `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Amount    float64   `json:"amount" example:"150000.50"`
	Term      int       `json:"term" example:"12"`
	Status    string    `json:"status" example:"NEW"`
	CreatedAt time.Time `json:"created_at" example:"2023-10-01T12:34:56Z"`
	UpdatedAt time.Time `json:"updated_at" example:"2023-10-01T12:34:56Z"`
}
