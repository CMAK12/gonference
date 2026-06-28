package config

type Config struct {
	WS   WS
	GRPC GRPC
}

type WS struct {
	Port int
}

type GRPC struct {
	Port int
}
