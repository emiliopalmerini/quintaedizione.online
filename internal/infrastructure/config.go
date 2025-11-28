package infrastructure

import (
	"os"
	"strconv"
	"time"
)

type Config struct {

	Port string
	Host string

	MongoURI     string
	DatabaseName string

	Environment string
	LogLevel    string

	ReadTimeout  time.Duration
	WriteTimeout time.Duration

	Pipeline PipelineConfig
}

type PipelineConfig struct {

	DefaultStages []string `json:"default_stages"`

	ErrorHandling string `json:"error_handling"`

	MaxParallelFiles int `json:"max_parallel_files"`

	EventBusBufferSize int `json:"event_bus_buffer_size"`

	ValidationEnabled bool `json:"validation_enabled"`

	TransformationEnabled bool `json:"transformation_enabled"`

	LogProgressEvents bool   `json:"log_progress_events"`
	LogLevel          string `json:"log_level"`
}

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

func getIntEnv(key string, fallback int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return fallback
}

func getBoolEnv(key string, fallback bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return fallback
}

func (pc PipelineConfig) ShouldStopOnError() bool {
	return pc.ErrorHandling == "stop"
}

func (pc PipelineConfig) ShouldContinueOnError() bool {
	return pc.ErrorHandling == "continue"
}

func (pc PipelineConfig) IsStageEnabled(stageName string) bool {
	for _, stage := range pc.DefaultStages {
		if stage == stageName {
			return true
		}
	}
	return false
}
