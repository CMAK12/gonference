// Package interceptors provides gRPC server interceptors shared by the
// services in this repository.
package interceptors

import (
	"context"
	"log/slog"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

// NewUnaryLogging returns an interceptor that logs the outcome and latency of
// every unary RPC. Failures are logged at error level, client errors at warn
// and successes at info.
func NewUnaryLogging(logger *slog.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		start := time.Now()

		resp, err := handler(ctx, req)

		logCall(ctx, logger, info.FullMethod, time.Since(start), err)

		return resp, err
	}
}

// NewStreamLogging returns the streaming counterpart of NewUnaryLogging. It
// logs once per stream, when the stream ends.
func NewStreamLogging(logger *slog.Logger) grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		start := time.Now()

		err := handler(srv, ss)

		logCall(ss.Context(), logger, info.FullMethod, time.Since(start), err)

		return err
	}
}

// logCall emits a single structured record describing a finished RPC.
func logCall(ctx context.Context, logger *slog.Logger, method string, duration time.Duration, err error) {
	code := status.Code(err)

	attrs := make([]slog.Attr, 0, 5)
	attrs = append(attrs,
		slog.String("method", method),
		slog.String("code", code.String()),
		slog.Duration("duration", duration),
	)

	if p, ok := peer.FromContext(ctx); ok {
		attrs = append(attrs, slog.String("peer", p.Addr.String()))
	}

	if err != nil {
		attrs = append(attrs, slog.String("error", err.Error()))
	}

	logger.LogAttrs(ctx, levelForCode(code), "grpc request", attrs...)
}

// levelForCode maps a status code to a log level so that client mistakes do not
// page whoever owns the server.
func levelForCode(code codes.Code) slog.Level {
	switch code {
	case codes.OK, codes.Canceled:
		return slog.LevelInfo
	case codes.InvalidArgument,
		codes.NotFound,
		codes.AlreadyExists,
		codes.PermissionDenied,
		codes.Unauthenticated,
		codes.FailedPrecondition,
		codes.OutOfRange,
		codes.ResourceExhausted:
		return slog.LevelWarn
	default:
		return slog.LevelError
	}
}
