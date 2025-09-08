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

	// Pipeline configuration
	Pipeline PipelineConfig
}

// PipelineConfig represents pipeline-specific configuration
type PipelineConfig struct {
	// Default stages to include in the pipeline
	DefaultStages []string `json:"default_stages"`

	// Error handling strategy: "continue" or "stop"
	ErrorHandling string `json:"error_handling"`

	// Maximum number of files to process in parallel
	MaxParallelFiles int `json:"max_parallel_files"`

	// EventBus configuration
	EventBusBufferSize int `json:"event_bus_buffer_size"`

	// Validation settings
	ValidationEnabled bool `json:"validation_enabled"`

	// Transformation settings
	TransformationEnabled bool `json:"transformation_enabled"`

	// Logging settings
	LogProgressEvents bool   `json:"log_progress_events"`
	LogLevel          string `json:"log_level"`
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
		Pipeline:     loadPipelineConfig(),
	}

	return config
}

// loadPipelineConfig loads pipeline-specific configuration
func loadPipelineConfig() PipelineConfig {
	return PipelineConfig{
		DefaultStages: []string{
			"file_reader",
			"content_parser",
			"validation",
			"transformation",
			"persistence",
			"error_handling",
		},
		ErrorHandling:         getEnv("PIPELINE_ERROR_HANDLING", "continue"),
		MaxParallelFiles:      getIntEnv("PIPELINE_MAX_PARALLEL_FILES", 5),
		EventBusBufferSize:    getIntEnv("PIPELINE_EVENT_BUS_BUFFER_SIZE", 1000),
		ValidationEnabled:     getBoolEnv("PIPELINE_VALIDATION_ENABLED", true),
		TransformationEnabled: getBoolEnv("PIPELINE_TRANSFORMATION_ENABLED", true),
		LogProgressEvents:     getBoolEnv("PIPELINE_LOG_PROGRESS_EVENTS", true),
		LogLevel:              getEnv("PIPELINE_LOG_LEVEL", "info"),
	}
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

// getIntEnv gets an integer from environment variable with fallback
func getIntEnv(key string, fallback int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return fallback
}

// getBoolEnv gets a boolean from environment variable with fallback
func getBoolEnv(key string, fallback bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return fallback
}

// Pipeline configuration methods

// ShouldStopOnError returns true if pipeline should stop on first error
func (pc PipelineConfig) ShouldStopOnError() bool {
	return pc.ErrorHandling == "stop"
}

// ShouldContinueOnError returns true if pipeline should continue on errors
func (pc PipelineConfig) ShouldContinueOnError() bool {
	return pc.ErrorHandling == "continue"
}

// IsStageEnabled checks if a specific stage is enabled in the configuration
func (pc PipelineConfig) IsStageEnabled(stageName string) bool {
	for _, stage := range pc.DefaultStages {
		if stage == stageName {
			return true
		}
	}
	return false
}
