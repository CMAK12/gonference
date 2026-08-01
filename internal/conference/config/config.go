package config

import "time"

type Config struct {
	GRPC         GRPC
	RelationalDB RelationalDB
	InMemoryDB   InMemoryDB
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

type RelationalDB struct {
	Driver string

	Host     string
	Port     string
	User     string
	Password string
	Database string
	Schema   string
	SSLMode  string

	MaxConnections  int32
	MinConnections  int32
	MaxConnLifetime time.Duration
	MaxConnIdleTime time.Duration
}

type InMemoryDB struct {
	Driver string

	Addr     string
	Password string
	Database int
}
