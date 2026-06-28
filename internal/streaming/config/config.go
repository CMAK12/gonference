package config

type Config struct {
	GRPC GRPC
}

type GRPC struct {
	Host string
	Port int
}
