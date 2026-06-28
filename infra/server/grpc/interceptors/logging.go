package interceptors

import (
	"context"
	"log/slog"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

func NewUnaryLogging(logger *slog.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		start := time.Now()

		resp, err = handler(ctx, req)

		duration := time.Since(start)
		code := status.Code(err)

		attrs := []any{
			"grpc.method", info.FullMethod,
			"code", code.String(),
			"duration_ms", duration.Milliseconds(),
		}

		if err != nil {
			attrs = append(attrs, "error", err.Error())
			logger.ErrorContext(ctx, "gRPC request failed", attrs...)
			return resp, err
		}

		logger.InfoContext(ctx, "gRPC request completed", attrs...)
		return resp, nil
	}
}
