package infrastructure

import (
	"os"
	"strconv"
	"time"
)

// Config represents application configuration
type Config struct {
	// Server configuration
	Port string
	Host string
	
	// MongoDB configuration
	MongoURI     string
	DatabaseName string
	
	// Application settings
	Environment string
	LogLevel    string
	
	// Timeouts
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

// LoadConfig loads configuration from environment variables
func LoadConfig() Config {
	config := Config{
		Port:         getEnv("PORT", "8000"),
		Host:         getEnv("HOST", "0.0.0.0"),
		MongoURI:     getEnv("MONGO_URI", "mongodb://localhost:27017"),
		DatabaseName: getEnv("DB_NAME", "dnd"),
		Environment:  getEnv("ENVIRONMENT", "development"),
		LogLevel:     getEnv("LOG_LEVEL", "info"),
		ReadTimeout:  getDurationEnv("READ_TIMEOUT", 15*time.Second),
		WriteTimeout: getDurationEnv("WRITE_TIMEOUT", 15*time.Second),
	}
	
	return config
}

// getEnv gets an environment variable with a fallback value
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// getDurationEnv gets a duration from environment variable with fallback
func getDurationEnv(key string, fallback time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if seconds, err := strconv.Atoi(value); err == nil {
			return time.Duration(seconds) * time.Second
		}
	}
	return fallback
}

// IsProduction returns true if running in production
func (c Config) IsProduction() bool {
	return c.Environment == "production"
}

// IsDevelopment returns true if running in development
func (c Config) IsDevelopment() bool {
	return c.Environment == "development"
}

// GetAddress returns the server address
func (c Config) GetAddress() string {
	return c.Host + ":" + c.Port
}