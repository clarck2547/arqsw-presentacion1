package config

import (
	"os"
	"strings"
)

type Config struct {
	SQLite struct {
		CommandDSN string
		QueryDSN   string
	}
	Kafka struct {
		Brokers []string
	}
	GRPC struct {
		Port string
	}
	HTTP struct {
		Port string
	}
}

func Load() *Config {
	return &Config{
		SQLite: struct {
			CommandDSN string
			QueryDSN   string
		}{
			CommandDSN: os.Getenv("COMMAND_DSN"),
			QueryDSN:   os.Getenv("QUERY_DSN"),
		},
		Kafka: struct {
			Brokers []string
		}{
			Brokers: strings.Split(os.Getenv("KAFKA_BROKERS"), ","),
		},
		GRPC: struct {
			Port string
		}{
			Port: os.Getenv("GRPC_PORT"),
		},
		HTTP: struct {
			Port string
		}{
			Port: os.Getenv("HTTP_PORT"),
		},
	}
}
