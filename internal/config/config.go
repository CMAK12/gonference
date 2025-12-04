package config

type Config struct {
	REST       REST
	AdminPanel AdminPanel
}

type REST struct {
	Port int
}

type AdminPanel struct {
	Port int
}
