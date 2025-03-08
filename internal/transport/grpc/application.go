package grpc

import (
	"context"
	"log"
	"time"

	"github.com/Andronzi/credit-origination/internal/domain"
	"github.com/Andronzi/credit-origination/internal/messaging"
	"github.com/Andronzi/credit-origination/internal/usecase"
	"github.com/Andronzi/credit-origination/pkg/grpc/credit"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type ApplicationServiceServer struct {
	credit.UnimplementedApplicationServiceServer
	getUC    *usecase.GetApplicationUseCase
	createUC *usecase.CreateApplicationUseCase
	listUC   *usecase.ListApplicationUseCase
	updateUC *usecase.UpdateApplicationUseCase
	deleteUC *usecase.DeleteApplicationUseCase
	producer *messaging.KafkaProducer
}

func ToDomainDecimal(d *credit.Decimal) decimal.Decimal {
	return decimal.New(d.Unscaled, d.Scale)
}

func ToProtoDecimal(d decimal.Decimal) *credit.Decimal {
	return &credit.Decimal{
		Unscaled: d.CoefficientInt64(),
		Scale:    int32(d.Exponent()),
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
	case credit.ApplicationStatus_APPLICATION:
		return domain.APPLICATION
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
	case domain.APPLICATION:
		return credit.ApplicationStatus_APPLICATION
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
	deleteUC *usecase.DeleteApplicationUseCase,
	producer *messaging.KafkaProducer,
) *ApplicationServiceServer {
	return &ApplicationServiceServer{
		getUC:    getUC,
		createUC: createUC,
		listUC:   listUC,
		updateUC: updateUC,
		deleteUC: deleteUC,
		producer: producer,
	}
}

func (s *ApplicationServiceServer) Create(ctx context.Context, req *credit.CreateApplicationRequest) (*credit.ApplicationResponse, error) {
	ID, err := StringToUUID(req.UserId)
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
		ID,
	)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if err := s.createUC.Execute(ctx, app); err != nil {
		return nil, status.Error(codes.Internal, "failed to create application")
	}

	event := messaging.StatusEvent{
		ApplicationID: app.ID.String(),
		EventType:     "AGREEMENT_CREATED",
		Timestamp:     time.Now().UnixMilli(),
		AgreementDetails: messaging.AgreementDetails{
			ApplicationID:      app.ID.String(),
			ClientID:           "client_id",
			DisbursementAmount: 1000,
			OriginationAmount:  1000,
			ToBankAccountID:    "account-id",
			Term:               10,
			Interest:           5,
			ProductCode:        "product-code-id",
			ProductVersion:     "product-version",
		},
	}

	if err := s.producer.SendStatusEvent(event); err != nil {
		log.Printf("Failed to send Kafka event: %v", err)
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
