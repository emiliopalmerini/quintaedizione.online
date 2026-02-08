package infrastructure

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Port string
	Host string

	Environment string
	LogLevel    string

	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

func LoadConfig() Config {
	if err := godotenv.Load(); err != nil {
		log.Println("Note: .env file not found, using environment variables")
	}

	config := Config{
		Port:         getEnv("PORT", "8000"),
		Host:         getEnv("HOST", "0.0.0.0"),
		Environment:  getEnv("ENVIRONMENT", "development"),
		LogLevel:     getEnv("LOG_LEVEL", "info"),
		ReadTimeout:  getDurationEnv("READ_TIMEOUT", 15*time.Second),
		WriteTimeout: getDurationEnv("WRITE_TIMEOUT", 15*time.Second),
	}

	return config
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getDurationEnv(key string, fallback time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if seconds, err := strconv.Atoi(value); err == nil {
			return time.Duration(seconds) * time.Second
		}
	}
	return fallback
}

func (c Config) IsProduction() bool {
	return c.Environment == "production"
}

func (c Config) IsDevelopment() bool {
	return c.Environment == "development"
}

func (c Config) GetAddress() string {
	return c.Host + ":" + c.Port
}
