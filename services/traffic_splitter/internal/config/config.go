package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	ExperimentManagerURL string
	ModelRegistryURL     string
	KafkaBrokers         string
	GRPCPort             string
	PollIntervalSeconds  string
}

func Load() *Config {
	_ = godotenv.Load()

	cfg := &Config{
		ExperimentManagerURL: getEnv("EXPERIMENT_SERVICE_URL", "http://localhost:8081"),
		ModelRegistryURL:     getEnv("MODEL_REGISTRY_URL", "http://localhost:8082"),
		KafkaBrokers:         getEnv("KAFKA_BOOTSTRAP_SERVERS", "localhost:9092"),
		GRPCPort:             getEnv("GRPC_PORT", "50051"),
		PollIntervalSeconds:  getEnv("POLL_INTERVAL_SECONDS", "30"),
	}

	validate(cfg)

	return cfg
}

func getEnv(key, def string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return def
}

func validate(cfg *Config) {
	if cfg.ExperimentManagerURL == "" {
		log.Fatal("EXPERIMENT_SERVICE_URL is required")
	}
	if cfg.ModelRegistryURL == "" {
		log.Fatal("MODEL_REGISTRY_URL is required")
	}
}
