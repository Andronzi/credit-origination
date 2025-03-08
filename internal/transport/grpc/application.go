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
	app, err := domain.NewCreditApplication(ToDomainDecimal(req.Amount), ToDomainDecimal(req.Interest), uint32(req.Term))
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
		Id:        app.ID.String(),
		Amount:    ToProtoDecimal(app.Amount),
		Term:      uint32(app.Term),
		Interest:  ToProtoDecimal(app.Interest),
		Status:    credit.ApplicationStatus(app.Status),
		CreatedAt: timestamppb.New(app.CreatedAt),
		UpdatedAt: timestamppb.New(app.UpdatedAt),
	}, nil
}

func (s *ApplicationServiceServer) List(ctx context.Context, req *credit.ListApplicationRequest) (*credit.ListApplicationResponse, error) {
	// TODO: Сделать качественнее маппинг
	protoStatuses := req.Status
	domainStatuses := make([]domain.ApplicationStatus, len(protoStatuses))

	for i, protoStatus := range protoStatuses {
		domainStatuses[i] = domain.ApplicationStatus(protoStatus)
	}

	result, err := s.listUC.Execute(ctx, domainStatuses, int(req.Page), int(req.PageSize))

	if err != nil {
		return nil, status.Error(codes.Internal, "failed to list applications")
	}

	var listApplicationResponses []*credit.ApplicationResponse

	for _, app := range result.Applications {
		listApplicationResponses = append(listApplicationResponses, &credit.ApplicationResponse{
			Id:        app.ID.String(),
			Amount:    ToProtoDecimal(app.Amount),
			Term:      uint32(app.Term),
			Interest:  ToProtoDecimal(app.Interest),
			Status:    credit.ApplicationStatus(app.Status),
			CreatedAt: timestamppb.New(app.CreatedAt),
			UpdatedAt: timestamppb.New(app.UpdatedAt),
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
		Id:        app.ID.String(),
		Amount:    ToProtoDecimal(app.Amount),
		Term:      uint32(app.Term),
		Interest:  ToProtoDecimal(app.Interest),
		Status:    credit.ApplicationStatus(app.Status),
		CreatedAt: timestamppb.New(app.CreatedAt),
		UpdatedAt: timestamppb.New(app.UpdatedAt),
	}, nil
}

func (s *ApplicationServiceServer) Update(ctx context.Context, req *credit.UpdateApplicationRequest) (*credit.ApplicationResponse, error) {
	now := time.Now()
	ID, err := StringToUUID(req.Id)
	if err != nil {
		return nil, err
	}

	app := &domain.CreditApplication{
		ID:        ID,
		Amount:    ToDomainDecimal(req.Amount),
		Term:      req.Term,
		Interest:  ToDomainDecimal(req.Interest),
		CreatedAt: now,
		UpdatedAt: now,
	}

	err = s.updateUC.Execute(ctx, app)

	if err != nil {
		return nil, status.Error(codes.Internal, "failed to load application")
	}

	return &credit.ApplicationResponse{
		Id:        app.ID.String(),
		Amount:    ToProtoDecimal(app.Amount),
		Term:      uint32(app.Term),
		Interest:  ToProtoDecimal(app.Interest),
		Status:    credit.ApplicationStatus(app.Status),
		CreatedAt: timestamppb.New(app.CreatedAt),
		UpdatedAt: timestamppb.New(app.UpdatedAt),
	}, nil
}

func (s *ApplicationServiceServer) Delete(ctx context.Context, req *credit.DeleteApplicationRequest) (*emptypb.Empty, error) {
	err := s.deleteUC.Execute(ctx, req.Id)

	if err != nil {
		return nil, status.Error(codes.Internal, "failed to delete application")
	}

	return &emptypb.Empty{}, nil
}
