package config

import "time"

type Config struct {
	ServiceName     string        `env:"SERVICE_NAME" envDefault:"service-template"`
	GRPCAddr        string        `env:"GRPC_ADDR" envDefault:":9090"`
	HTTPAddr        string        `env:"HTTP_ADDR" envDefault:":8081"`
	LogLevel        string        `env:"LOG_LEVEL" envDefault:"info"`
	ShutdownTimeout time.Duration `env:"SHUTDOWN_TIMEOUT" envDefault:"30s"`
}
