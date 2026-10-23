package config

import "time"

type Config struct {
	GRPC       GRPC
	Relay      Relay
	Signaling  Signaling
	InMemoryDB InMemoryDB
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

type Relay struct {
	AdvertiseAddr string
}

type Signaling struct {
	MaxResolveRetries int
	RelayOwnershipTTL time.Duration
}

type InMemoryDB struct {
	Addr     string
	Password string
	Database int
}
