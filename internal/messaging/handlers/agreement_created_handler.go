package handlers

import (
	"context"

	"github.com/Andronzi/credit-origination/internal/domain"
	"github.com/Andronzi/credit-origination/internal/messaging"
	"github.com/Andronzi/credit-origination/internal/usecase"
	"github.com/Andronzi/credit-origination/pkg/logger"
	"github.com/IBM/sarama"
	"github.com/google/uuid"
	"github.com/linkedin/goavro/v2"
	"go.uber.org/zap"
)

type AgreementCreatedHandler struct {
	updateStatusUC *usecase.UpdateStatusUseCase
	codec          *goavro.Codec
}

func NewAgreementCreatedHandler(uc *usecase.UpdateStatusUseCase, schema string) (*AgreementCreatedHandler, error) {
	codec, err := goavro.NewCodec(schema)
	if err != nil {
		return nil, err
	}
	return &AgreementCreatedHandler{
		updateStatusUC: uc,
		codec:          codec,
	}, nil
}

func (h *AgreementCreatedHandler) Handle(message *sarama.ConsumerMessage) error {
	logger.Logger.Info("Start handle message in AgreementCreatedHandler")

	if len(message.Value) < 5 {
		logger.Logger.Error("Invalid message length")
		return nil
	}
	avroData := message.Value[5:]

	native, _, err := h.codec.NativeFromBinary(avroData)
	if err != nil {
		logger.Logger.Error("Failed to decode avro", zap.Error(err))
		return err
	}

	data, ok := native.(map[string]interface{})
	if !ok {
		logger.Logger.Error("Unexpected message format")
		return nil
	}

	applicationID, ok := data["application_id"].(string)
	if !ok {
		logger.Logger.Error("Invalid application_id in message")
		return nil
	}

	appID, err := uuid.Parse(applicationID)
	if err != nil {
		logger.Logger.Error("Invalid application ID", zap.String("ID", applicationID))
		return err
	}

	ctx := context.Background()
	if err := h.updateStatusUC.Execute(ctx, appID, domain.SCORING); err != nil {
		logger.Logger.Error("Failed to update status to SCORING", zap.Error(err))
		return err
	}

	return nil
}

var _ messaging.MessageHandler = (*AgreementCreatedHandler)(nil)
