package middleware

import (
	"context"
	"math/rand/v2"
	"time"

	"github.com/Andronzi/credit-origination/pkg/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func ErrorInjectionInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		now := time.Now()
		errorProbability := 0.5
		if now.Minute()%2 == 0 {
			errorProbability = 0.9
		}

		if rand.Float64() < errorProbability {
			logger.Logger.Warn("Injecting error",
				zap.Time("time", now),
				zap.Float64("probability", errorProbability),
			)
			return nil, status.Error(codes.Internal, "random error injection")
		}
		return handler(ctx, req)
	}
}
