package parsers

import "fmt"

type Logger interface {
	Info(msg string, args ...interface{})
	Error(msg string, args ...interface{})
	Debug(msg string, args ...interface{})
}

type NoOpLogger struct{}

func (l *NoOpLogger) Info(msg string, args ...interface{}) {}

func (l *NoOpLogger) Error(msg string, args ...interface{}) {}

func (l *NoOpLogger) Debug(msg string, args ...interface{}) {}

type ConsoleLogger struct {
	prefix string
}

func NewConsoleLogger(prefix string) *ConsoleLogger {
	return &ConsoleLogger{prefix: prefix}
}

func (l *ConsoleLogger) Info(msg string, args ...interface{}) {
	fmt.Printf("[%s] INFO: %s\n", l.prefix, fmt.Sprintf(msg, args...))
}

func (l *ConsoleLogger) Error(msg string, args ...interface{}) {
	fmt.Printf("[%s] ERROR: %s\n", l.prefix, fmt.Sprintf(msg, args...))
}

func (l *ConsoleLogger) Debug(msg string, args ...interface{}) {
	fmt.Printf("[%s] DEBUG: %s\n", l.prefix, fmt.Sprintf(msg, args...))
}
