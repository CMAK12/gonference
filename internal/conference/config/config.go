package config

import (
	"github.com/CMAK12/gonference/infra/db/in-memory/dragonfly"
	"github.com/CMAK12/gonference/infra/db/relational/postgres"
)

// Config is the full configuration for the conference service.
type Config struct {
	GRPC      GRPC
	Storage   Storage
	Postgres  postgres.Config
	Dragonfly dragonfly.Config
}

type GRPC struct {
	Address string
}

// Storage selects which persistence backend the service uses.
type Storage struct {
	// Driver is "postgres" (relational) or "dragonfly" (in-memory).
	Driver string
}
