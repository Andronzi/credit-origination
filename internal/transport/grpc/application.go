package grpc

import (
	"context"
	"math"
	"time"

	"github.com/Andronzi/credit-origination/internal/domain"
	"github.com/Andronzi/credit-origination/internal/messaging"
	"github.com/Andronzi/credit-origination/internal/usecase"
	"github.com/Andronzi/credit-origination/pkg/grpc/credit"
	"github.com/Andronzi/credit-origination/pkg/logger"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type ApplicationServiceServer struct {
	credit.UnimplementedApplicationServiceServer
	getUC          *usecase.GetApplicationUseCase
	createUC       *usecase.CreateApplicationUseCase
	listUC         *usecase.ListApplicationUseCase
	updateUC       *usecase.UpdateApplicationUseCase
	updateStatusUC *usecase.UpdateStatusUseCase
	deleteUC       *usecase.DeleteApplicationUseCase
	producer       *messaging.KafkaProducer
}

func ToDomainDecimal(d *credit.Decimal) decimal.Decimal {
	return decimal.New(d.Unscaled, -d.Scale)
}

func ToProtoDecimal(d decimal.Decimal) *credit.Decimal {
	return &credit.Decimal{
		Unscaled: d.CoefficientInt64(),
		Scale:    int32(math.Abs(float64(d.Exponent()))),
	}
}

func StringToUUID(idStr string) (uuid.UUID, error) {
	id, err := uuid.Parse(idStr)
	if err != nil {
		return uuid.UUID{}, status.Error(codes.Internal, "id has wrong format")
	}
	return id, nil
}

func MapGRPCStatusToDomain(grpcStatus credit.ApplicationStatus) domain.ApplicationStatus {
	switch grpcStatus {
	case credit.ApplicationStatus_DRAFT:
		return domain.DRAFT
	case credit.ApplicationStatus_APPLICATION_CREATED:
		return domain.APPLICATION_CREATED
	case credit.ApplicationStatus_APPLICATION_AGREEMENT_CREATED:
		return domain.APPLICATION_AGREEMENT_CREATED
	case credit.ApplicationStatus_SCORING:
		return domain.SCORING
	case credit.ApplicationStatus_EMPLOYMENT_CHECK:
		return domain.EMPLOYMENT_CHECK
	case credit.ApplicationStatus_APPROVED:
		return domain.APPROVED
	case credit.ApplicationStatus_REJECTED:
		return domain.REJECTED
	default:
		return domain.DRAFT
	}
}

func MapDomainStatusToGRPC(domainStatus domain.ApplicationStatus) credit.ApplicationStatus {
	switch domainStatus {
	case domain.DRAFT:
		return credit.ApplicationStatus_DRAFT
	case domain.APPLICATION_CREATED:
		return credit.ApplicationStatus_APPLICATION_CREATED
	case domain.APPLICATION_AGREEMENT_CREATED:
		return credit.ApplicationStatus_APPLICATION_AGREEMENT_CREATED
	case domain.SCORING:
		return credit.ApplicationStatus_SCORING
	case domain.EMPLOYMENT_CHECK:
		return credit.ApplicationStatus_EMPLOYMENT_CHECK
	case domain.APPROVED:
		return credit.ApplicationStatus_APPROVED
	case domain.REJECTED:
		return credit.ApplicationStatus_REJECTED
	default:
		return credit.ApplicationStatus_DRAFT
	}
}

func NewCreateApplicationServer(
	getUC *usecase.GetApplicationUseCase,
	createUC *usecase.CreateApplicationUseCase,
	listUC *usecase.ListApplicationUseCase,
	updateUC *usecase.UpdateApplicationUseCase,
	updateStatusUC *usecase.UpdateStatusUseCase,
	deleteUC *usecase.DeleteApplicationUseCase,
	producer *messaging.KafkaProducer,
) *ApplicationServiceServer {
	return &ApplicationServiceServer{
		getUC:          getUC,
		createUC:       createUC,
		listUC:         listUC,
		updateUC:       updateUC,
		updateStatusUC: updateStatusUC,
		deleteUC:       deleteUC,
		producer:       producer,
	}
}

func (s *ApplicationServiceServer) Create(ctx context.Context, req *credit.CreateApplicationRequest) (*credit.ApplicationResponse, error) {
	logger.Logger.Info("Received create application request",
		zap.String("service", "ApplicationServiceServer.Create"),
		zap.String("user_id", req.UserId),
		zap.String("to_bank_account_id", req.ToBankAccountId),
		zap.Any("request", req),
	)

	userID, err := StringToUUID(req.UserId)
	if err != nil {
		return nil, err
	}
	app, err := domain.NewCreditApplication(
		ToDomainDecimal(req.DisbursementAmount),
		ToDomainDecimal(req.OriginationAmount),
		uuid.MustParse(req.ToBankAccountId),
		uint32(req.Term),
		ToDomainDecimal(req.Interest),
		req.ProductCode,
		req.ProductVersion,
		userID,
		MapGRPCStatusToDomain(req.Status),
	)
	if err != nil {
		logger.Logger.Error("Failed to create new credit application",
			zap.String("user_id", req.UserId),
			zap.Error(err),
		)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	logger.Logger.Info("Successfully created credit application",
		zap.String("app_id", app.ID.String()),
	)

	if err := s.createUC.Execute(ctx, app); err != nil {
		logger.Logger.Error("createUC execution failed",
			zap.String("app_id", app.ID.String()),
			zap.Error(err),
		)
		return nil, status.Error(codes.Internal, "failed to create application")
	}
	logger.Logger.Info("createUC executed successfully",
		zap.String("app_id", app.ID.String()),
	)

	if err := s.updateStatusUC.Execute(ctx, app.ID, domain.APPLICATION_AGREEMENT_CREATED); err != nil {
		logger.Logger.Error("updateStatusUC execution failed",
			zap.String("app_id", app.ID.String()),
			zap.Error(err),
		)
		return nil, status.Error(codes.Internal, "status update failed")
	}
	logger.Logger.Info("Application status updated",
		zap.String("app_id", app.ID.String()),
	)

	resp := &credit.ApplicationResponse{
		Id:                 app.ID.String(),
		UserId:             app.UserID.String(),
		DisbursementAmount: ToProtoDecimal(app.DisbursementAmount),
		OriginationAmount:  ToProtoDecimal(app.OriginationAmount),
		ToBankAccountId:    app.ToBankAccountID.String(),
		Term:               uint32(app.Term),
		Interest:           ToProtoDecimal(app.Interest),
		Status:             MapDomainStatusToGRPC(app.Status),
		ProductCode:        app.ProductCode,
		ProductVersion:     app.ProductVersion,
		CreatedAt:          timestamppb.New(app.CreatedAt),
		UpdatedAt:          timestamppb.New(app.UpdatedAt),
	}
	logger.Logger.Info("Sending response for create application",
		zap.String("app_id", app.ID.String()),
	)

	return resp, nil
}

func (s *ApplicationServiceServer) List(ctx context.Context, req *credit.ListApplicationRequest) (*credit.ListApplicationResponse, error) {
	// TODO: Сделать качественнее маппинг
	protoStatuses := req.Status
	domainStatuses := make([]domain.ApplicationStatus, len(protoStatuses))

	for i, protoStatus := range protoStatuses {
		domainStatuses[i] = MapGRPCStatusToDomain(protoStatus)
	}

	result, err := s.listUC.Execute(ctx, domainStatuses, int(req.Page), int(req.PageSize))

	if err != nil {
		return nil, status.Error(codes.Internal, "failed to list applications")
	}

	var listApplicationResponses []*credit.ApplicationResponse

	for _, app := range result.Applications {
		listApplicationResponses = append(listApplicationResponses, &credit.ApplicationResponse{
			Id:                 app.ID.String(),
			UserId:             app.UserID.String(),
			DisbursementAmount: ToProtoDecimal(app.DisbursementAmount),
			OriginationAmount:  ToProtoDecimal(app.OriginationAmount),
			ToBankAccountId:    app.ToBankAccountID.String(),
			Term:               uint32(app.Term),
			Interest:           ToProtoDecimal(app.Interest),
			Status:             MapDomainStatusToGRPC(app.Status),
			ProductCode:        app.ProductCode,
			ProductVersion:     app.ProductVersion,
			CreatedAt:          timestamppb.New(app.CreatedAt),
			UpdatedAt:          timestamppb.New(app.UpdatedAt),
		})
	}

	return &credit.ListApplicationResponse{
		Applications: listApplicationResponses,
		TotalCount:   uint32(result.TotalCount),
		Page:         req.Page,
		PageSize:     uint32(req.PageSize),
		TotalPages:   uint32(result.TotalPages),
	}, nil
}

func (s *ApplicationServiceServer) Get(ctx context.Context, req *credit.GetApplicationRequest) (*credit.ApplicationResponse, error) {
	app, err := s.getUC.Execute(ctx, req.Id)

	if err != nil {
		return nil, status.Error(codes.Internal, "failed to load application")
	}

	return &credit.ApplicationResponse{
		Id:                 app.ID.String(),
		UserId:             app.UserID.String(),
		DisbursementAmount: ToProtoDecimal(app.DisbursementAmount),
		OriginationAmount:  ToProtoDecimal(app.OriginationAmount),
		ToBankAccountId:    app.ToBankAccountID.String(),
		Term:               uint32(app.Term),
		Interest:           ToProtoDecimal(app.Interest),
		Status:             MapDomainStatusToGRPC(app.Status),
		ProductCode:        app.ProductCode,
		ProductVersion:     app.ProductVersion,
		CreatedAt:          timestamppb.New(app.CreatedAt),
		UpdatedAt:          timestamppb.New(app.UpdatedAt),
	}, nil
}

func (s *ApplicationServiceServer) Update(ctx context.Context, req *credit.UpdateApplicationRequest) (*credit.ApplicationResponse, error) {
	now := time.Now()
	ID, err := StringToUUID(req.Id)
	if err != nil {
		return nil, err
	}
	UserID, err := StringToUUID(req.UserId)
	if err != nil {
		return nil, err
	}
	ToBankAccountId, err := StringToUUID(req.ToBankAccountId)
	if err != nil {
		return nil, err
	}

	app := &domain.CreditApplication{
		ID:                 ID,
		UserID:             UserID,
		DisbursementAmount: ToDomainDecimal(req.DisbursementAmount),
		OriginationAmount:  ToDomainDecimal(req.OriginationAmount),
		ToBankAccountID:    ToBankAccountId,
		Term:               uint32(req.Term),
		Interest:           ToDomainDecimal(req.Interest),
		Status:             MapGRPCStatusToDomain(req.Status),
		ProductCode:        req.ProductCode,
		ProductVersion:     req.ProductVersion,
		UpdatedAt:          now,
	}

	err = s.updateUC.Execute(ctx, app)

	if err != nil {
		return nil, status.Error(codes.Internal, "failed to load application")
	}

	return &credit.ApplicationResponse{
		Id:                 app.ID.String(),
		UserId:             app.UserID.String(),
		DisbursementAmount: ToProtoDecimal(app.DisbursementAmount),
		OriginationAmount:  ToProtoDecimal(app.OriginationAmount),
		ToBankAccountId:    app.ToBankAccountID.String(),
		Term:               uint32(app.Term),
		Interest:           ToProtoDecimal(app.Interest),
		Status:             MapDomainStatusToGRPC(app.Status),
		ProductCode:        app.ProductCode,
		ProductVersion:     app.ProductVersion,
		CreatedAt:          timestamppb.New(app.CreatedAt),
		UpdatedAt:          timestamppb.New(app.UpdatedAt),
	}, nil
}

func (s *ApplicationServiceServer) Delete(ctx context.Context, req *credit.DeleteApplicationRequest) (*emptypb.Empty, error) {
	err := s.deleteUC.Execute(ctx, req.Id)

	if err != nil {
		return nil, status.Error(codes.Internal, "failed to delete application")
	}

	return &emptypb.Empty{}, nil
}
