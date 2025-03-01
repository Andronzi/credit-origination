package grpc

import (
	"context"

	"github.com/Andronzi/credit-origination/internal/domain"
	"github.com/Andronzi/credit-origination/internal/usecase"
	"github.com/Andronzi/credit-origination/pkg/grpc/credit"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type ApplicationServiceServer struct {
	credit.UnimplementedApplicationServiceServer
	createUC *usecase.CreateApplicationUseCase
	listUC   *usecase.ListApplicationUseCase
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
	app, err := domain.NewCreditApplication(req.Amount, int(req.Term))

	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if err := s.createUC.Execute(ctx, app); err != nil {
		return nil, status.Error(codes.Internal, "failed to create application")
	}

	return &credit.ApplicationResponse{
		Id:        app.ID.String(),
		Amount:    app.Amount,
		Term:      int32(app.Term),
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
			Amount:    app.Amount,
			Term:      int32(app.Term),
			Interest:  app.Interest,
			Status:    credit.ApplicationStatus(app.Status),
			CreatedAt: timestamppb.New(app.CreatedAt),
			UpdatedAt: timestamppb.New(app.UpdatedAt),
		})
	}

	return &credit.ListApplicationResponse{
		Applications: listApplicationResponses,
		TotalCount:   int32(result.TotalCount),
		Page:         req.Page,
		PageSize:     int32(req.PageSize),
		TotalPages:   int32(result.TotalPages),
	}, nil
}
