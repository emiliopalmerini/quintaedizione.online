package parsers

import "fmt"

// Logger interface for parsing operations
type Logger interface {
	Info(msg string, args ...interface{})
	Error(msg string, args ...interface{})
	Debug(msg string, args ...interface{})
}

// NoOpLogger provides a logger that discards all output
type NoOpLogger struct{}

// Info implements Logger interface
func (l *NoOpLogger) Info(msg string, args ...interface{}) {}

// Error implements Logger interface
func (l *NoOpLogger) Error(msg string, args ...interface{}) {}

// Debug implements Logger interface
func (l *NoOpLogger) Debug(msg string, args ...interface{}) {}

// ConsoleLogger provides console-based logging with prefix
type ConsoleLogger struct {
	prefix string
}

// NewConsoleLogger creates a new console logger with the given prefix
func NewConsoleLogger(prefix string) *ConsoleLogger {
	return &ConsoleLogger{prefix: prefix}
}

// Info logs an info message to console
func (l *ConsoleLogger) Info(msg string, args ...interface{}) {
	fmt.Printf("[%s] INFO: %s\n", l.prefix, fmt.Sprintf(msg, args...))
}

// Error logs an error message to console
func (l *ConsoleLogger) Error(msg string, args ...interface{}) {
	fmt.Printf("[%s] ERROR: %s\n", l.prefix, fmt.Sprintf(msg, args...))
}

// Debug logs a debug message to console
func (l *ConsoleLogger) Debug(msg string, args ...interface{}) {
	fmt.Printf("[%s] DEBUG: %s\n", l.prefix, fmt.Sprintf(msg, args...))
}
