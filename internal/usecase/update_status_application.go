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

func MapDomainStatusToAvro(status domain.ApplicationStatus) string {
	if status == domain.APPROVED {
		return "DISBURSEMENT_PROCESSED"
	} else {
		return string(status)
	}
}

func CreatePaymentDate(status domain.ApplicationStatus) *int64 {
	logger.Logger.Info("CreatePaymentDate", zap.String("status", string(status)))
	if status == domain.APPROVED {
		now := time.Now()
		startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		unixStartOfDay := startOfDay.Unix()
		logger.Logger.Info("CreatePaymentDate", zap.String("tinme", string(unixStartOfDay)))
		return &unixStartOfDay
	} else {
		return nil
	}
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
		EventType:     MapDomainStatusToAvro(app.Status),
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
			PaymentDate:        CreatePaymentDate(app.Status),
		},
	}

	logger.Logger.Info("createStatusEvent: event generated",
		zap.String("app_id", app.ID.String()),
		zap.Any("event", event),
	)
	return event
}
