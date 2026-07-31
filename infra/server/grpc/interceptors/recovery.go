package interceptors

import (
	"context"
	"fmt"
	"log/slog"
	"runtime"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// stackBufSize bounds the stack trace captured on panic.
const stackBufSize = 8 << 10 // 8 KiB

// NewUnaryRecovery returns an interceptor that converts a panic in a unary
// handler into a codes.Internal error so one bad request cannot take the
// process down. The panic value and stack are logged; the client is told
// nothing beyond "internal error".
func NewUnaryRecovery(logger *slog.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		defer func() {
			if r := recover(); r != nil {
				logPanic(ctx, logger, info.FullMethod, r)

				resp, err = nil, status.Error(codes.Internal, "internal error")
			}
		}()

		return handler(ctx, req)
	}
}

// NewStreamRecovery returns the streaming counterpart of NewUnaryRecovery.
func NewStreamRecovery(logger *slog.Logger) grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
		defer func() {
			if r := recover(); r != nil {
				logPanic(ss.Context(), logger, info.FullMethod, r)

				err = status.Error(codes.Internal, "internal error")
			}
		}()

		return handler(srv, ss)
	}
}

// logPanic records a recovered panic together with its stack trace.
func logPanic(ctx context.Context, logger *slog.Logger, method string, recovered any) {
	buf := make([]byte, stackBufSize)
	stack := buf[:runtime.Stack(buf, false)]

	logger.LogAttrs(ctx, slog.LevelError, "grpc handler panic",
		slog.String("method", method),
		slog.String("panic", fmt.Sprint(recovered)),
		slog.String("stack", string(stack)),
	)
}
