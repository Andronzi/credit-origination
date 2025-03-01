package request

type CreditApplicationCreateRequest struct {
	Amount float64 `json:"amount" example:"150000.50" binding:"required,gt=0"`
	Term   int     `json:"term" example:"12" binding:"required,gt=0"`
}
