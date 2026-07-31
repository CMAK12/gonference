package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"sync/atomic"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"

	"github.com/CMAK12/gonference/infra/server/grpc/interceptors"
)

const (
	DefaultMaxRecvMsgSize = 4 << 20 // 4 MiB
	DefaultMaxSendMsgSize = 4 << 20 // 4 MiB

	DefaultConnectionTimeout = 20 * time.Second

	DefaultMaxConnectionIdle     = 15 * time.Minute
	DefaultKeepaliveTime         = 2 * time.Minute
	DefaultKeepaliveTimeout      = 20 * time.Second
	DefaultMinClientPingInterval = time.Minute

	DefaultShutdownTimeout = 10 * time.Second

	loggerComponent = "grpc-server"
)

var ErrAlreadyServing = errors.New("grpc server is already serving")

type Server struct {
	*grpc.Server

	log      *slog.Logger
	listener net.Listener

	shutdownTimeout time.Duration
	served          atomic.Bool
}

type config struct {
	log   *slog.Logger
	creds credentials.TransportCredentials

	maxRecvMsgSize int
	maxSendMsgSize int

	connectionTimeout time.Duration
	shutdownTimeout   time.Duration

	keepalive   keepalive.ServerParameters
	enforcement keepalive.EnforcementPolicy

	unary  []grpc.UnaryServerInterceptor
	stream []grpc.StreamServerInterceptor
	extra  []grpc.ServerOption

	reflection bool
}

type Option func(*config)

func WithLogger(log *slog.Logger) Option {
	return func(c *config) {
		if log != nil {
			c.log = log
		}
	}
}

func WithCredentials(creds credentials.TransportCredentials) Option {
	return func(c *config) { c.creds = creds }
}

func WithMessageSizes(recv, send int) Option {
	return func(c *config) {
		if recv > 0 {
			c.maxRecvMsgSize = recv
		}
		if send > 0 {
			c.maxSendMsgSize = send
		}
	}
}

func WithConnectionTimeout(d time.Duration) Option {
	return func(c *config) { c.connectionTimeout = d }
}

func WithShutdownTimeout(d time.Duration) Option {
	return func(c *config) { c.shutdownTimeout = d }
}

func WithKeepalive(params keepalive.ServerParameters) Option {
	return func(c *config) { c.keepalive = params }
}

func WithConnectionAge(age, grace time.Duration) Option {
	return func(c *config) {
		c.keepalive.MaxConnectionAge = age
		c.keepalive.MaxConnectionAgeGrace = grace
	}
}

func WithKeepaliveEnforcement(policy keepalive.EnforcementPolicy) Option {
	return func(c *config) { c.enforcement = policy }
}

func WithUnaryInterceptors(in ...grpc.UnaryServerInterceptor) Option {
	return func(c *config) { c.unary = append(c.unary, in...) }
}

func WithStreamInterceptors(in ...grpc.StreamServerInterceptor) Option {
	return func(c *config) { c.stream = append(c.stream, in...) }
}

func WithReflection(enabled bool) Option {
	return func(c *config) { c.reflection = enabled }
}

func WithServerOptions(opts ...grpc.ServerOption) Option {
	return func(c *config) { c.extra = append(c.extra, opts...) }
}

func New(addr string, opts ...Option) (*Server, error) {
	if addr == "" {
		return nil, errors.New("address is required")
	}

	cfg := config{
		log:               slog.Default(),
		maxRecvMsgSize:    DefaultMaxRecvMsgSize,
		maxSendMsgSize:    DefaultMaxSendMsgSize,
		connectionTimeout: DefaultConnectionTimeout,
		shutdownTimeout:   DefaultShutdownTimeout,
		keepalive: keepalive.ServerParameters{
			MaxConnectionIdle: DefaultMaxConnectionIdle,
			Time:              DefaultKeepaliveTime,
			Timeout:           DefaultKeepaliveTimeout,
		},
		enforcement: keepalive.EnforcementPolicy{
			MinTime:             DefaultMinClientPingInterval,
			PermitWithoutStream: true,
		},
	}

	for _, opt := range opts {
		opt(&cfg)
	}

	log := cfg.log.With(slog.String("component", loggerComponent))

	serverOpts := []grpc.ServerOption{
		grpc.MaxRecvMsgSize(cfg.maxRecvMsgSize),
		grpc.MaxSendMsgSize(cfg.maxSendMsgSize),
		grpc.ConnectionTimeout(cfg.connectionTimeout),
		grpc.KeepaliveParams(cfg.keepalive),
		grpc.KeepaliveEnforcementPolicy(cfg.enforcement),
		grpc.ChainUnaryInterceptor(append(
			[]grpc.UnaryServerInterceptor{
				interceptors.NewUnaryRecovery(log),
				interceptors.NewUnaryLogging(log),
			},
			cfg.unary...,
		)...),
		grpc.ChainStreamInterceptor(append(
			[]grpc.StreamServerInterceptor{
				interceptors.NewStreamRecovery(log),
				interceptors.NewStreamLogging(log),
			},
			cfg.stream...,
		)...),
	}

	if cfg.creds != nil {
		serverOpts = append(serverOpts, grpc.Creds(cfg.creds))
	}

	serverOpts = append(serverOpts, cfg.extra...)

	srv := grpc.NewServer(serverOpts...)

	if cfg.reflection {
		reflection.Register(srv)
	}

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		srv.Stop()

		return nil, err
	}

	return &Server{
		Server:          srv,
		log:             log.With(slog.String("addr", listener.Addr().String())),
		listener:        listener,
		shutdownTimeout: cfg.shutdownTimeout,
	}, nil
}

func (s *Server) Addr() string {
	return s.listener.Addr().String()
}

func (s *Server) Serve() error {
	if !s.served.CompareAndSwap(false, true) {
		return ErrAlreadyServing
	}

	s.log.Info("gRPC server listening")

	if err := s.Server.Serve(s.listener); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
		return fmt.Errorf("serve grpc on %s: %w", s.Addr(), err)
	}

	s.log.Info("gRPC server stopped serving")

	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.log.Info("gRPC server shutting down")

	done := make(chan struct{})

	go func() {
		defer close(done)

		s.Server.GracefulStop()
	}()

	select {
	case <-done:
		s.releaseListener()
		s.log.Info("gRPC server shut down gracefully")

		return nil
	case <-ctx.Done():
		s.log.Warn("gRPC graceful shutdown deadline exceeded, forcing stop",
			slog.String("error", ctx.Err().Error()),
		)
		s.Server.Stop()
		<-done
		s.releaseListener()

		return fmt.Errorf("shutdown grpc on %s: %w", s.Addr(), ctx.Err())
	}
}

func (s *Server) GracefulStop() {
	ctx, cancel := context.WithTimeout(context.Background(), s.shutdownTimeout)
	defer cancel()

	_ = s.Shutdown(ctx)
}

func (s *Server) Close() error {
	s.Server.Stop()

	if err := s.closeListener(); err != nil {
		return fmt.Errorf("close grpc listener on %s: %w", s.Addr(), err)
	}

	return nil
}

func (s *Server) releaseListener() {
	if s.served.Load() {
		return
	}

	if err := s.closeListener(); err != nil {
		s.log.Warn("failed to close gRPC listener", slog.String("error", err.Error()))
	}
}

func (s *Server) closeListener() error {
	if err := s.listener.Close(); err != nil && !errors.Is(err, net.ErrClosed) {
		return err
	}

	return nil
}
