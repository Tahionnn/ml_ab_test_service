package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port                 string
	UserServiceURL       string
	ExperimentManagerURL string
	ModelRegistryURL     string
	SplitterAddress      string
	KafkaBrokers         []string
	PredictionTopic      string
}

func Load() *Config {
	_ = godotenv.Load()

	cfg := &Config{
		Port:                 getEnv("PORT", "8080"),
		UserServiceURL:       getEnv("USER_SERVICE_URL", "http://user_service:8000"),
		ExperimentManagerURL: getEnv("EXPERIMENT_SERVICE_URL", "http://experiment_manager:8001"),
		ModelRegistryURL:     getEnv("MODEL_REGISTRY_URL", "http://model_registry:8002"),
		SplitterAddress:      getEnv("TRAFFIC_SPLITTER_ADDRESS", "splitter:50051"),
		KafkaBrokers:         []string{getEnv("KAFKA_BOOTSTRAP_SERVERS", "kafka:9092")},
		PredictionTopic:      getEnv("PREDICTION_TOPIC", "prediction.events"),
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
