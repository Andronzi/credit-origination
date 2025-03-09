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

type ScoringHandler struct {
	updateStatusUC *usecase.UpdateStatusUseCase
	codec          *goavro.Codec
}

func NewScoringHandler(uc *usecase.UpdateStatusUseCase, schema string) (*ScoringHandler, error) {
	codec, err := goavro.NewCodec(schema)
	if err != nil {
		return nil, err
	}
	return &ScoringHandler{
		updateStatusUC: uc,
		codec:          codec,
	}, nil
}

func (h *ScoringHandler) Handle(message *sarama.ConsumerMessage) error {
	logger.Logger.Info("Start handle message in ScoringHandler")

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
	if err := h.updateStatusUC.Execute(ctx, appID, domain.APPROVED); err != nil {
		logger.Logger.Error("Failed to update status to APPROVED", zap.Error(err))
		return err
	}

	return nil
}

var _ messaging.MessageHandler = (*ScoringHandler)(nil)
