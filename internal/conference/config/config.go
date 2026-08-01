package config

import (
	"time"

	"github.com/CMAK12/gonference/infra/db/in-memory/dragonfly"
	"github.com/CMAK12/gonference/infra/db/relational/postgres"
)

type Config struct {
	GRPC      GRPC
	Storage   Storage
	Postgres  postgres.Config
	Dragonfly dragonfly.Config
}

type GRPC struct {
	Addr string

	MaxRecvMsgSize int
	MaxSendMsgSize int

	ConnectionTimeout time.Duration
	ShutdownTimeout   time.Duration

	MaxConnectionIdle     time.Duration
	KeepaliveTime         time.Duration
	KeepaliveTimeout      time.Duration
	MinClientPingInterval time.Duration

	Reflection bool
}

type Storage struct {
	Driver string
}
