package handlers

import (
	"github.com/Andronzi/credit-origination/internal/domain"
	"github.com/Andronzi/credit-origination/internal/usecase"
	"github.com/gin-gonic/gin"
)

type CreditHandler struct {
	createUC *usecase.CreateApplicationUseCase
	getUC    *usecase.GetApplicationUseCase
}

func NewCreateCreditHandler(
	createUC *usecase.CreateApplicationUseCase,
	getUC    *usecase.GetApplicationUseCase
) *CreditHandler {
	return &CreditHandler{
		createUC: createUC
		getUC: getUC,
	}
}

func (h *CreditHandler) CreateApplication(c *gin.Context) {
	var request struct {
		UserID string  `json:"user_id"`
		Amount float64 `json:"amount"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(400, gin.H{"error": "invalid request"})
		return
	}

	app := &domain.CreditApplication{
		UserID: request.UserID,
		Amount: request.Amount,
		Status: domain.StatusNew,
	}

	if err := h.createUC.Execute(c.Request.Context(), app); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(201, app)
}
