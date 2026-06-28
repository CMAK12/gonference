package server

import (
	"log/slog"
	"net"

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
		grpc.ChainUnaryInterceptor(),
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
