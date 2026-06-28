package config

type Config struct {
	GRPC GRPC
}

type GRPC struct {
	Address string
}
