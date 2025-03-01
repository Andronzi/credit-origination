package grpc

import (
	"context"

	"github.com/Andronzi/credit-origination/internal/domain"
	"github.com/Andronzi/credit-origination/internal/usecase"
	"github.com/Andronzi/credit-origination/pkg/grpc/credit"
	"github.com/shopspring/decimal"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type ApplicationServiceServer struct {
	credit.UnimplementedApplicationServiceServer
	createUC *usecase.CreateApplicationUseCase
	listUC   *usecase.ListApplicationUseCase
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

func NewCreateApplicationServer(
	createUC *usecase.CreateApplicationUseCase,
	listUC *usecase.ListApplicationUseCase,
) *ApplicationServiceServer {
	return &ApplicationServiceServer{
		createUC: createUC,
		listUC:   listUC,
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
	result, err := s.listUC.Execute(ctx, domain.ApplicationStatus(req.Status), int(req.Page), int(req.PageSize))

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
