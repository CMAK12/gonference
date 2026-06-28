package server

import (
	"log/slog"
	"net"

	"github.com/CMAK12/gonference/infra/server/grpc/interceptors"
	"google.golang.org/grpc"
)

type Server struct {
	*grpc.Server

	log *slog.Logger
	net net.Listener

	Addr string
}

func New(addr string) (*Server, error) {
	s := grpc.NewServer(
		grpc.Creds(nil),
		grpc.ChainUnaryInterceptor(
			interceptors.NewUnaryLogging(slog.Default().With("component", "grpc-server")),
		),
	)

	l, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	return &Server{
		Server: s,
		Addr:   addr,
		net:    l,
		log:    slog.Default().With("component", "grpc-server"),
	}, nil
}

func (s *Server) Serve() error {
	s.log.Info("Starting gRPC server", slog.String("addr", s.Addr))
	return s.Server.Serve(s.net)
}

func (s *Server) GracefulStop() {
	s.log.Info("Gracefully stopping gRPC server", slog.String("addr", s.Addr))
	s.Server.GracefulStop()
	s.log.Info("gRPC server stopped", slog.String("addr", s.Addr))
}
