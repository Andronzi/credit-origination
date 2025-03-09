package usecase

import (
	"context"
	"time"

	"github.com/Andronzi/credit-origination/internal/domain"
	"github.com/Andronzi/credit-origination/internal/messaging"
	"github.com/Andronzi/credit-origination/pkg/logger"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type UpdateStatusUseCase struct {
	repo     domain.CreditRepository
	producer *messaging.KafkaProducer
}

func NewUpdateStatusUseCase(
	repo domain.CreditRepository,
	producer *messaging.KafkaProducer,
) *UpdateStatusUseCase {
	return &UpdateStatusUseCase{repo, producer}
}

func (uc *UpdateStatusUseCase) Execute(ctx context.Context, appID uuid.UUID, newStatus domain.ApplicationStatus) error {
	logger.Logger.Info("UpdateStatusUseCase.Execute started",
		zap.String("app_id", appID.String()),
		zap.String("new_status", string(newStatus)),
	)

	app, err := uc.repo.FindByID(ctx, appID.String())
	if err != nil {
		logger.Logger.Error("Failed to find application",
			zap.String("app_id", appID.String()),
			zap.Error(err),
		)
		return err
	}
	logger.Logger.Info("Application found",
		zap.String("app_id", app.ID.String()),
		zap.String("current_status", string(app.Status)),
	)

	if err := app.ChangeStatus(newStatus); err != nil {
		logger.Logger.Error("Failed to change application status",
			zap.String("app_id", app.ID.String()),
			zap.String("new_status", string(newStatus)),
			zap.Error(err),
		)
		return err
	}
	logger.Logger.Info("Application status changed",
		zap.String("app_id", app.ID.String()),
		zap.String("new_status", string(newStatus)),
	)

	if err := uc.repo.Update(ctx, app); err != nil {
		logger.Logger.Error("Failed to update application in repository",
			zap.String("app_id", app.ID.String()),
			zap.Error(err),
		)
		return err
	}
	logger.Logger.Info("Application updated in repository",
		zap.String("app_id", app.ID.String()),
	)

	event := uc.createStatusEvent(app)
	logger.Logger.Info("Status event created",
		zap.String("app_id", app.ID.String()),
		zap.Any("event", event),
	)
	if err := uc.producer.SendStatusEvent(event); err != nil {
		logger.Logger.Error("Failed to send status event",
			zap.String("app_id", app.ID.String()),
			zap.Error(err),
		)
		return err
	}
	logger.Logger.Info("Status event sent successfully",
		zap.String("app_id", app.ID.String()),
	)

	return nil
}

func (uc *UpdateStatusUseCase) createStatusEvent(app *domain.CreditApplication) messaging.ApplicationStatusEvent {
	event := messaging.ApplicationStatusEvent{
		ApplicationID: app.ID.String(),
		EventType:     string(app.Status),
		Timestamp:     time.Now().UnixMilli(),
		AgreementDetails: messaging.AgreementDetails{
			ApplicationID:      app.ID.String(),
			ClientID:           app.UserID.String(),
			DisbursementAmount: app.DisbursementAmount.IntPart(),
			OriginationAmount:  app.OriginationAmount.IntPart(),
			ToBankAccountID:    app.ToBankAccountID.String(),
			Term:               int32(app.Term),
			Interest:           app.Interest.IntPart(),
			ProductCode:        app.ProductCode,
			ProductVersion:     app.ProductVersion,
		},
	}

	logger.Logger.Info("createStatusEvent: event generated",
		zap.String("app_id", app.ID.String()),
		zap.Any("event", event),
	)
	return event
}
