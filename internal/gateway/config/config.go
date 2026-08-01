package config

import "time"

type Config struct {
	GRPC   GRPC
	Client Client
}

type (
	GRPC struct {
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

	Client struct {
		Conference struct {
			Addr string
		}

		Streaming struct {
			Addr string
		}
	}
)
